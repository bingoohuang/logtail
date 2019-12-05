package util

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	"github.com/bingoohuang/gou/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// DeclareLogPFlags declares the log pflags.
func DeclareLogPFlags() {
	pflag.StringP("loglevel", "", "info", "debug/info/warn/error")
	pflag.StringP("logdir", "", "var/logs", "log dir")
}

// SetupLog setup log parameters.
func SetupLog() io.Writer {
	logdir := viper.GetString("logdir")
	if logdir != "" {
		if err := os.MkdirAll(logdir, os.ModePerm); err != nil {
			logrus.Panicf("failed to create %s error %v\n", logdir, err)
			return os.Stdout
		}

		loglevel := viper.GetString("loglevel")

		return lo.InitLogger(loglevel, logdir, filepath.Base(os.Args[0])+".log")
	}

	logrus.SetLevel(logrus.DebugLevel)

	return os.Stdout
}
