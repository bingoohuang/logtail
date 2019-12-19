// +build !solaris

package tail

import (
	"io"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"

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

	FromBeginning bool `pflag:"Read file from beginning"`
	Pipe          bool `pflag:"Whether file is a named pipe"`

	liner   Liner
	tailers map[string]*tail.Tail

	WatchMethod      string   `pflag:"Method used to watch for file updates(inotify/poll), default inotify"`
	OffsetSavePrefix string   `pflag:"Offset save file prefix in in ~, default logtail"`
	Files            []string `pflag:"Files to tail"`
}

// NewTail create a new tail.
func NewTail(liner Liner) *Tail {
	return &Tail{FromBeginning: false, liner: liner, OffsetSavePrefix: "logtail"}
}

// every 10 seconds
func (t *Tail) Gather() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// always start from the beginning of files that appear while we're running
	t.tailNewFiles(true)
}

// sampleConfig =
//  # files to tail.
//  # These accept standard unix glob matching rules, but with the addition of
//  # ** as a "super asterisk". ie:
//  #   "/var/log/**.log"  -> recursively find all .log files in /var/log
//  #   "/var/log/*/*.log" -> find all .log files with a parent dir in /var/log
//  #   "/var/log/apache.log" -> just tail the apache log file
//  #
//  # See https://github.com/gobwas/glob for more examples
//  #
//  Files = ["/var/mymetrics.out"]
//  # Read file from beginning.
//  FromBeginning = false
//  # Whether file is a named pipe
//  Pipe = false
//  # Method used to watch for file updates.  Can be either "inotify" or "poll".
//  # WatchMethod = "inotify"

// Start starts a tail go routine.
func (t *Tail) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.tailers = make(map[string]*tail.Tail)
	t.tailNewFiles(t.FromBeginning)

	if len(t.tailers) == 0 {
		logrus.Warnf("no files to tail")
	}
}

func (t *Tail) tailNewFiles(fromBeginning bool) {
	// Create a "tailer" for each file
	for _, filepath := range t.Files {
		if src, err := homedir.Expand(filepath); err == nil {
			filepath = src
		}

		g, err := globpath.Compile(filepath)
		if err != nil {
			logrus.Errorf("Glob %q failed to compile: %s", filepath, err.Error())
		}

		matches := g.Match()
		logrus.Infof("tailer filepath %s matches %v", filepath, matches)

		for _, file := range matches {
			logrus.Infof("tailer filepath %s match %s", filepath, file)

			if _, ok := t.tailers[file]; ok {
				continue // we're already tailing this file
			}

			t.createTailer(file, fromBeginning)
		}
	}
}

func (t *Tail) createTailer(file string, fromBeginning bool) {
	var offset *tail.SeekInfo
	if !t.Pipe {
		offset = &tail.SeekInfo{Whence: io.SeekStart, Offset: 0}
		if !fromBeginning {
			offset, _ = ReadTailFileOffset(t.OffsetSavePrefix, file, offset)
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

	var offset int64 = -1

ForLoop:
	for {
		select {
		case <-ticker.C:
			offsetChanged := false
			offset, offsetChanged = SaveTailerOffset(t.OffsetSavePrefix, tailer, offset)
			if offsetChanged {
				logrus.Debugf("Recording offset %d for %q", offset, tailer.Filename)
			}
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
		if !t.Pipe && !t.FromBeginning {
			// store offset for resume
			if offset, err := tailer.Tell(); err == nil {
				logrus.Debugf("Recording offset %d for %q", offset, tailer.Filename)
			} else {
				logrus.Errorf("Recording offset for %q: %s", tailer.Filename, err.Error())
			}

			SaveTailerOffset(t.OffsetSavePrefix, tailer, -1)
		}

		if err := tailer.Stop(); err != nil {
			logrus.Errorf("Stopping tail on %q: %s", tailer.Filename, err.Error())
		}
	}

	t.wg.Wait()
}
