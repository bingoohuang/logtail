package tail

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/influxdata/tail"
	homedir "github.com/mitchellh/go-homedir"
)

// GetFileIno 获得文件的Inode值。这样，在文件被重建时（文件名不变），可以使用Inode来确定文件是否被重新创建
func GetFileIno(file string) (uint64, error) {
	fileinfo, err := os.Stat(file)
	if err != nil {
		return 0, err
	}

	stat, ok := fileinfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, errors.New(file + " is not a syscall.Stat_t")
	}

	return stat.Ino, nil
}

// ReadTailFileOffset 读取tail文件对应的.offset文件记录的偏移位置
func ReadTailFileOffset(prefix, file string, fallback *tail.SeekInfo) (*tail.SeekInfo, error) {
	offset, err := ioutil.ReadFile(getTailerOffsetFileName(prefix, file))
	if err != nil {
		return fallback, err
	}

	off, err := strconv.ParseInt(string(offset), 10, 64)
	if err != nil {
		return fallback, err
	}

	return &tail.SeekInfo{Whence: io.SeekStart, Offset: off}, nil
}

// SaveTailerOffset 保存tail文件读取的位置到对应的.offset文件
func SaveTailerOffset(prefix string, tailer *tail.Tail, lastOffset int64) (offset int64, changed bool) {
	offset, _ = tailer.Tell()
	if lastOffset == offset {
		return offset, false
	}

	b := []byte(strconv.FormatInt(offset, 10))
	_ = ioutil.WriteFile(getTailerOffsetFileName(prefix, tailer.Filename), b, 0644)

	return offset, true
}

// ClearTailerOffset 删除tail文件对应的.offset文件
func ClearTailerOffset(prefix string, tailer *tail.Tail) {
	_ = os.Remove(getTailerOffsetFileName(prefix, tailer.Filename))
}

// getTailerOffsetFileName 获得tail文件对应的.offset文件名称
func getTailerOffsetFileName(prefix, file string) string {
	executable, _ := os.Executable()
	dir, _ := homedir.Expand("~/." + prefix + "-" + filepath.Base(executable) + "/tailoffset/")
	_ = os.MkdirAll(dir, os.ModePerm)

	ino, err := GetFileIno(file)
	if err != nil {
		f := strings.ReplaceAll(file, string(os.PathSeparator), "_")
		return filepath.Join(dir, f+".offset")
	}

	return filepath.Join(dir, "inode"+fmt.Sprintf("%d", ino)+".offset")
}
