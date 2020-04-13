# logtail

[![Travis CI](https://img.shields.io/travis/bingoohuang/logtail/master.svg?style=flat-square)](https://travis-ci.com/bingoohuang/logtail)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/bingoohuang/logtail/blob/master/LICENSE.md)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/bingoohuang/logtail)
[![Coverage Status](http://codecov.io/github/bingoohuang/logtail/coverage.svg?branch=master)](http://codecov.io/github/bingoohuang/logtail?branch=master)
[![goreport](https://www.goreportcard.com/badge/github.com/bingoohuang/logtail)](https://www.goreportcard.com/repo

tail log and do something

## Build

`go install ./...`

## 日志行提取规则

1. PreMatches 预匹配，含有预匹配字符串的日志行入选
2. Splitter 分隔符切割，如果指定了分隔符，那么先用分隔符切割
3. CaptureSplitSeq 选取切割子串序号（从1开始）
4. CaptureReq 如果设置了正则匹配，使用正则匹配，否则跳到6号锚定点匹配
5. CaptureGroup 正则匹配后选取子捕获分组（默认0，表示选取正则匹配的0号分组），跳到7
6. AnchorStart AnchorEnd 使用锚定点开始和结束匹配
7. CaptureCut 最终修剪（如果设置了）

例如：

```toml
# 预匹配(子串包含)
PreMatches = ["[End]", "AuthService"]
# 分隔符切割，如果指定了分隔符，那么先用分隔符切割
Splitter = "^_^"
# 选取切割子串序号（从1开始）
CaptureSplitSeq = 2
# 正则匹配
CaptureReg   = ""
# 正则匹配后选取子捕获分组
CaptureGroup = 0
# 起始锚点
AnchorStart = "["
# 结束锚点
AnchorEnd  = ""
# 最终修剪
CaptureCut = "0:-1"
```

对于日志行：

```
2020-04-13 15:05:55,955  INFO 16376 --- [http-nio-12-exec-8] [72] c.o.MonitorLogger            : AuthService.customerVerify(..)[End]:158^_^AuthService.customerVerify(..)=[{"data":"Ynzaa==","signAlgo":"HmacSHA256","appId":"61578c46","version":"1.0","deviceId":"DEV_1db8f","algo":"SHA256withRSA"}]^_^{"message":"成功","status":200}^_^10^_^false
```

1. 首先预匹配发现，改行包含`[End]`和`AuthService`，所以入选
2. 使用`^_^`分割，得到以下子串
    1. `2020-04-13 15:05:55,955  INFO 16376 --- [http-nio-12-exec-8] [72] c.o.MonitorLogger            : AuthService.customerVerify(..)[End]:158`
    2. `AuthService.customerVerify(..)=[{"data":"Ynzaa==","signAlgo":"HmacSHA256","appId":"61578c46","version":"1.0","deviceId":"DEV_1db8f","algo":"SHA256withRSA"}]`
    3. `{"message":"成功","status":200}`
    4. `10`
    5. `false`
3. 选取切割子串序号上面`2`的子串
4. CaptureReg 正则匹配没有配置，跳过
5. CaptureGroup 跳过
6. AnchorStart="[",AnchorEnd=""，得到子串：`{"data":"Ynzaa==","signAlgo":"HmacSHA256","appId":"61578c46","version":"1.0","deviceId":"DEV_1db8f","algo":"SHA256withRSA"}]`
7. 使用`CaptureCut = "0:-1"`修剪掉最后一个字符，得到最终字符串`{"data":"Ynzaa==","signAlgo":"HmacSHA256","appId":"61578c46","version":"1.0","deviceId":"DEV_1db8f","algo":"SHA256withRSA"}`

## Post重放请求响应报文响应配置

```toml
# 比较响应-切分后取第几个子串(1开始)
CmpRspSplitSeq  = 3
# 比较响应-匹配正则表达式
CmpRspCaptureReg  = ""
# 比较响应-捕获组序号
CmpRspCaptureGroup = 0
# 比较响应-起始锚点
CmpRspAnchorStart = ""
# 比较响应-终止锚点
CmpRspAnchorEnd  = ""
# 比较响应-切割，eg: 切除首尾字符 1:-1，切除尾部1一个字符:-1
CmpRspCut   = ""
# 比较响应-比较通过日志文件名，不配置不打印
CmdRspOKLog = "ok.log"
# 比较响应-比较失败日志文件名，不配置，输出到stderr
CmdRspBadLog = "bad.log"
```

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

# 预匹配（子串包含）
PreMatches = ["[End]"]
# POST URL
PostURL  = "http://127.0.0.1:8812"

# 切割分隔符
Splitter = "^_^"

# 切分后取第几个子串(1开始)
CaptureSplitSeq = 2

# 匹配正则，优先级高
CaptureReq  = ''''''
# 正则匹配，捕获组序号
CaptureGroup = 0

# 在Capture为空时，使用锚点定位

# 起始锚点
AnchorStart = '''customerVerify(..)=['''
# 终止锚点
AnchorEnd = ""
# 切除尾部1一个字符:
CaptureCut=":-1"

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
