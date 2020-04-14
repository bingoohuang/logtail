# logtail

[![Travis CI](https://img.shields.io/travis/bingoohuang/logtail/master.svg?style=flat-square)](https://travis-ci.com/bingoohuang/logtail)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/bingoohuang/logtail/blob/master/LICENSE.md)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/bingoohuang/logtail)
[![Coverage Status](http://codecov.io/github/bingoohuang/logtail/coverage.svg?branch=master)](http://codecov.io/github/bingoohuang/logtail?branch=master)
[![goreport](https://www.goreportcard.com/badge/github.com/bingoohuang/logtail)](https://www.goreportcard.com/report/github.com/bingoohuang/logtail)

tail log and do something

## Build

`go install ./...`

## 日志行提取规则

使用管道处理方式：
例如：`contains customerVerify [End] | split by=^_^ keeps=1 | anchor start=[ | cut :-1`
表示使用grep进行匹配，然后分割后取第1(0-based)部分，然后再锚点`[`后的部分，最后裁剪掉最后一个字符。

目前支持的处理器：

1. `contains sub1 sub2 ...` 处理行必须包含sub1,sub2
1. `split [by={byValue}] [keeps={keepsValue}]` 按byValue切分，如果by参数没有指定，则按空白字符切分，然后保留索引值(0-based)为keepsValue的部分，例如`keeps=1`表示保留第1部分，`keeps=1,3`表示保留第1，3部分，不指定默认保留第0部分
1. `anchor [start={startValue}] [end={endValue}] [includeStart=yes] [includeEnd=yes]` 按锚点取值，例如`anchor start={ end=}` 则表示取`{`和`}`中间的值
1. `cut from:to` 表示切割，`cut 1:-1` 切割掉首位各1位字符
1. `join by={byValue}` 合并
1. `trim` 去除两端空白
1. `reg expr [group1 group2 ...]` 使用正则提取值，指定提取组号的子匹配部分，默认提取组号为0的子匹配
1. `grep expr1 [expr2 expr3 ...]` 处理行必须符合正则expr1, expr2, expr3 ...


## Usage

demo [config.toml](testdata/cnf1.toml)

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

# 日志行捕获表达式
Capture = "grep customerVerify [End] | split by=^_^ keeps=1 | anchor start=[ | cut :-1"
# POST URL
PostURL  = "http://127.0.0.1:8812"
# 期望POST返回体捕获表达式
ExpectRsp = "split by=^_^ keeps=2"
# 比较响应-比较通过日志文件名
RspOKLog  = ""
# 比较响应-比较失败日志文件名
RspFailLog = "fail.log"
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
$ logtail -h                                                                                                                                                                         [Tue Apr 14 09:32:34 2020]
  Usage of logtail:
        --capture string            捕获表达式
    -c, --cnf string                cnf file path
        --expectRsp string          期待响应表达式
        --files string              Files to tail
        --fromBeginning             Read file from beginning
    -i, --init                      init to create template config file and ctl.sh
        --logdir string             log dir (default "var/logs")
        --loglevel string           debug/info/warn/error (default "info")
        --logrus                    enable logrus (default true)
        --offsetSavePrefix string   Offset save file prefix in in ~, default logtail
        --pipe                      Whether file is a named pipe
        --postUrl string            POST URL
        --rspFailLog string         比较响应-比较失败日志文件名
        --rspOklog string           比较响应-比较通过日志文件名
    -v, --version                   show version
        --watchMethod string        Method used to watch for file updates(inotify/poll), default inotify
  pflag: help requested
```

```bash
logtail --files ./MSSM-Auth.log --fromBeginning=true --logrus=false --capture="contains customerVerify [End] | split by=^_^ keeps=1 | anchor start=[ | cut :-1"
```


## References

1. http://httpbin.org/ This is a simple HTTP Request & Response Service
1. https://jsonplaceholder.typicode.com A placeholder website to invoke REST APIs
