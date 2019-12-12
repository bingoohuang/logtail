package main

import (
	"fmt"

	"github.com/bingoohuang/gou/lo"
	"github.com/bingoohuang/gou/sy"

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
	lo.DeclareLogPFlags()

	if err := cnf.ParsePflags("LOGTAIL"); err != nil {
		panic(err)
	}

	if *ver {
		fmt.Println("Version: v0.2.2")
		return
	}

	gsutil.Ipo(*ipo)

	cnf.LoadByPflag(&tailer, &linerPost)

	lo.SetupLog()

	if err := linerPost.Setup(); err != nil {
		panic(err)
	}

	logrus.Infof("tailer config %s", enc.JSONCompact(tailer))
	logrus.Infof("linerPost config %s", enc.JSONCompact(linerPost))

	go tailer.Start()

	done := make(chan bool, 1)

	sy.AwaitingSignal(done, func() { tailer.Stop() })

	<-done
	logrus.Infof("exiting")
}
