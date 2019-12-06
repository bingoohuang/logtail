// +build !solaris

package tail

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/logtail/internal/globpath"
	"github.com/influxdata/tail"
	"github.com/sirupsen/logrus"
)

// Liner processes a line of tailer.
type Liner interface {
	ProcessLine(tailer *tail.Tail, line string, firstLine bool) error
}

// Tail tail log files.
type Tail struct {
	mu sync.Mutex
	wg sync.WaitGroup

	FromBeginning bool
	Pipe          bool

	liner   Liner
	tailers map[string]*tail.Tail

	WatchMethod string // default "inotify"
	Files       []string
}

// NewTail create a new tail.
func NewTail(liner Liner) *Tail {
	return &Tail{FromBeginning: false, liner: liner}
}

// sampleConfig =
//  ## files to tail.
//  ## These accept standard unix glob matching rules, but with the addition of
//  ## ** as a "super asterisk". ie:
//  ##   "/var/log/**.log"  -> recursively find all .log files in /var/log
//  ##   "/var/log/*/*.log" -> find all .log files with a parent dir in /var/log
//  ##   "/var/log/apache.log" -> just tail the apache log file
//  ##
//  ## See https://github.com/gobwas/glob for more examples
//  ##
//  Files = ["/var/mymetrics.out"]
//  ## Whether file is a named pipe
//  Pipe = false
//  ## Method used to watch for file updates.  Can be either "inotify" or "poll".
//  # WatchMethod = "inotify"

// Start starts a tail go routine.
func (t *Tail) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.tailers = make(map[string]*tail.Tail)

	t.tailNewFiles(t.FromBeginning)
}

func (t *Tail) tailNewFiles(fromBeginning bool) {
	var seek *tail.SeekInfo
	if !t.Pipe && !t.FromBeginning {
		seek = &tail.SeekInfo{
			Whence: io.SeekEnd,
			Offset: 0,
		}
	}
	// Create a "tailer" for each file
	for _, filepath := range t.Files {
		fmt.Printf("tailer filepath:%s \n", filepath)
		g, err := globpath.Compile(filepath)
		if err != nil {
			logrus.Errorf("Glob %q failed to compile: %s", filepath, err.Error())
		}

		for _, file := range g.Match() {
			fmt.Printf("tailer filepath %s match:%s \n", filepath, file)
			if _, ok := t.tailers[file]; ok {
				continue // we're already tailing this file
			}

			t.createTailer(file, seek)
		}
	}
}

func (t *Tail) createTailer(file string, seek *tail.SeekInfo) {
	offset := ReadTailFileOffset(file, seek)
	if offset == nil && t.FromBeginning {
		offset = &tail.SeekInfo{
			Whence: io.SeekCurrent,
			Offset: 0,
		}
	}
	tailConfig := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  offset,
		MustExist: true,
		Poll:      t.WatchMethod == "poll",
		Pipe:      t.Pipe,
		Logger:    tail.DiscardingLogger,
	}

	tailer, err := tail.TailFile(file, tailConfig)
	if err != nil {
		logrus.Debugf("Failed to open file (%s): %v", file, err)
		return
	}

	logrus.Debugf("Tail added for %q", file)

	// create a goroutine for each "tailer"
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()
		t.receiver(tailer)
	}()

	t.tailers[tailer.Filename] = tailer
}

// Receiver is launched as a goroutine to continuously watch a tailed logfile
// for changes, parse any incoming msgs, and add to the accumulator.
func (t *Tail) receiver(tailer *tail.Tail) {
	firstLine := true

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

ForLoop:
	for {
		select {
		case <-ticker.C:
			offset := SaveTailerOffset(tailer)
			logrus.Debugf("SaveTailerOffset %s, offset:%d", tailer.Filename, offset)
		case line, ok := <-tailer.Lines:
			if !ok {
				break ForLoop
			}

			if line.Err != nil {
				logrus.Errorf("Tailing %q: %s", tailer.Filename, line.Err.Error())
				continue
			}

			// Fix up files with Windows line endings.
			text := strings.TrimRight(line.Text, "\r")

			if t.liner != nil {
				if err := t.liner.ProcessLine(tailer, text, firstLine); err != nil {
					logrus.Errorf("Malformed log line in %q: [%q]: %s", tailer.Filename, line.Text, err.Error())
				}
			}

			firstLine = false
		}
	}

	logrus.Debugf("Tail removed for %q", tailer.Filename)

	if err := tailer.Err(); err != nil {
		logrus.Errorf("Tailing %q: %s", tailer.Filename, err.Error())
	}
}

// Stop stops the tail goroutine.
func (t *Tail) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, tailer := range t.tailers {
		if !t.Pipe {
			// store offset for resume
			if offset, err := tailer.Tell(); err == nil {
				logrus.Debugf("Recording offset %d for %q", offset, tailer.Filename)
			} else {
				logrus.Errorf("Recording offset for %q: %s", tailer.Filename, err.Error())
			}
		}

		SaveTailerOffset(tailer)

		if err := tailer.Stop(); err != nil {
			logrus.Errorf("Stopping tail on %q: %s", tailer.Filename, err.Error())
		}
	}

	t.wg.Wait()
}
