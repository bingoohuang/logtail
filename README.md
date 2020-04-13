# logtail

[![Travis CI](https://img.shields.io/travis/bingoohuang/logtail/master.svg?style=flat-square)](https://travis-ci.com/bingoohuang/logtail)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/bingoohuang/logtail/blob/master/LICENSE.md)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/bingoohuang/logtail)
[![Coverage Status](http://codecov.io/github/bingoohuang/logtail/coverage.svg?branch=master)](http://codecov.io/github/bingoohuang/logtail?branch=master)
[![goreport](https://www.goreportcard.com/badge/github.com/bingoohuang/logtail)](https://www.goreportcard.com/repo

tail log and do something

## Build

`go install ./...`

## Usage

demo [config.toml](testdata/cnf.toml)

```toml
# files to tail.
# These accept standard unix glob matching rules, but with the addition of
# ** as a "super asterisk". ie:
#   "/var/log/**.log"  -> recursively find all .log files in /var/log
#   "/var/log/*/*.log" -> find all .log files with a parent dir in /var/log
#   "/var/log/apache.log" -> just tail the apache log file
#
# See https://github.com/gobwas/glob for more examples
Files = ["testdata/*.log"]
# Read file from beginning.
FromBeginning = false
# Whether file is a named pipe
Pipe = false

# 前置匹配（子串包含）
Matches = ["[End]"]
# POST URL
PostURL  = "http://127.0.0.1:8812"

# 匹配正则，优先级高
Capture  = ''''''
# 正则匹配，捕获组序号
CaptureGroup = 0

# 在Capture为空时，使用锚点定位

# 起始锚点
AnchorStart = '''customerVerify(..)=['''
# 终止锚点
AnchorEnd = ''']^_^'''

```

Start the POST mock server

```bash
$ logtailmockserver
2019/12/05 20:39:57 start mock server on :8812
```

Start the logtail (Notice: remove offset cache file by `rm -fr ~/.logtail*` to start to tail from beginning of log files)

```bash
$ logtail -i
./ctl created!
./config.toml created!
$ rm -fr ~/.logtail*
$ ./ctl start
logtail now is running already, pid=86594
$ ./ctl status                                                                                                                                                                                              ➜  logtail git:(master) ✗ ./ctl status
logtail started, pid=86594
```

```bash
$ logtail -h
Usage of logtail:
      --anchorEnd string          终止锚点（在Capture为空时有效）
      --anchorStart string        起始锚点（在Capture为空时有效）
      --capture string            匹配正则
      --captureGroup int          捕获组序号
  -c, --cnf string                cnf file path
      --files string              Files to tail
      --fromBeginning             Read file from beginning
  -i, --init                      init to create template config file and ctl.sh
      --logdir string             log dir (default "var/logs")
      --loglevel string           debug/info/warn/error (default "info")
      --logrus                    enable logrus (default true)
      --matches string            前置匹配（子串包含）
      --offsetSavePrefix string   Offset save file prefix in in ~, default logtail
      --pipe                      Whether file is a named pipe
      --postURL string            POST URL
  -v, --version                   show version
      --watchMethod string        Method used to watch for file updates(inotify/poll), default inotify
pflag: help requested
```


## References

1. http://httpbin.org/ This is a simple HTTP Request & Response Service
1. https://jsonplaceholder.typicode.com A placeholder website to invoke REST APIs
