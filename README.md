# logtail
tail log and do something

## Build

`go install ./...`

## Usage

demo [config.toml](testdata/config.toml)

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

# 锚点
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
       --anchorEnd string
       --anchorStart string
       --capture string
       --captureGroup int
   -c, --cnf string           cnf file path
       --files string
   -i, --init                 init to create template config file and ctl.sh
       --logdir string        log dir (default "var/logs")
       --loglevel string      debug/info/warn/error (default "info")
       --matches string
       --pipe
       --postURL string
   -v, --version              show version
       --watchMethod string
 pflag: help requested
```
