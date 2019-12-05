package tail

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/araddon/gou"
	"github.com/influxdata/tail"
	"github.com/stretchr/testify/assert"
)

func TestTailFileOffset(t *testing.T) {
	gou.SetupLogging("debug")

	a := assert.New(t)

	file := "TestTailFileOffset.log"
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	a.Nil(err, "打开文件成功")

	tailer, err := tail.TailFile(file,
		tail.Config{
			ReOpen:    true,
			Follow:    true,
			Location:  ReadTailFileOffset(file),
			MustExist: true,
		})
	a.Nil(err, "Tail文件成功")

	msg := "HelloWorld" + fmt.Sprintf("%d", rand.Intn(30))

	_, _ = f.WriteString(msg + "\n")

	select {
	case line := <-tailer.Lines:
		a.Equal(line.Text, msg)
	case <-time.After(10 * time.Second):
		a.Fail("没有读到消息")
	}

	SaveTailerOffset(tailer)
}
