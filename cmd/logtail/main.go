package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bingoohuang/gou/cnf"
	"github.com/bingoohuang/gou/enc"

	"github.com/sirupsen/logrus"

	gsutil "github.com/bingoohuang/gostarter/util"
	"github.com/bingoohuang/logtail/liner"
	"github.com/bingoohuang/logtail/tail"
	"github.com/spf13/pflag"

	_ "github.com/bingoohuang/logtail/statiq"
)

func main() {
	ver := pflag.BoolP("version", "v", false, "show version")
	ipo := pflag.BoolP("init", "i", false, "init to create template config file and ctl.sh")

	var linerPost liner.Post

	tailer := tail.NewTail(&linerPost)

	cnf.DeclarePflags()
	cnf.DeclarePflagsByStruct(tailer, linerPost)
	gsutil.DeclareLogPFlags()

	if err := cnf.ParsePflags("LOGTAIL"); err != nil {
		panic(err)
	}

	if *ver {
		fmt.Println("Version: v0.2.2")
		return
	}

	gsutil.Ipo(*ipo)

	cnf.LoadByPflag(&tailer, &linerPost)

	gsutil.SetupLog()

	if err := linerPost.Setup(); err != nil {
		panic(err)
	}

	logrus.Infof("tailer config %s", enc.JSONCompact(tailer))
	logrus.Infof("linerPost config %s", enc.JSONCompact(linerPost))

	go tailer.Start()

	done := make(chan bool, 1)

	startSignal(done, func() { tailer.Stop() })

	<-done
	logrus.Infof("exiting")
}

func startSignal(done chan bool, f func()) {
	// Go signal notification works by sending `os.Signal`
	// values on a channel. We'll create a channel to
	// receive these notifications (we'll also make one to
	// notify us when the program can exit).
	sigs := make(chan os.Signal, 1)
	// `signal.Notify` registers the given channel to
	// receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// This goroutine executes a blocking receive for
	// signals. When it gets one it'll print it out
	// and then notify the program that it can finish.
	go func() {
		sig := <-sigs
		logrus.Infof("got signal %v", sig)
		f()
		done <- true
	}()

	// The program will wait here until it gets the
	// expected signal (as indicated by the goroutine
	// above sending a value on `done`) and then exit.
	logrus.Infof("awaiting signal interrupt(ctl+c) or terminated(by kill)")
}
