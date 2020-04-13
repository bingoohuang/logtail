package capture

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/bingoohuang/gou/str"
)

// DistractConfig defines the configuration to distract a sub string.
type DistractConfig struct {
	SplitSeq int

	Capture      string
	CaptureGroup int

	AnchorStart string
	AnchorEnd   string

	Cut string

	captureReg *regexp.Regexp

	cutFrom, cutTo int
}

// IsEmpty tells whether the config is wholly empty or not.
func (p *DistractConfig) IsEmpty() bool {
	return p.SplitSeq == 0 && p.Capture == "" && p.AnchorStart == "" && p.AnchorEnd == ""
}

func (p *DistractConfig) setup() error {
	var err error

	if p.Capture != "" {
		if p.captureReg, err = regexp.Compile(p.Capture); err != nil {
			return fmt.Errorf("compile regex %s  error %w", p.Capture, err)
		}
	}

	// nolint gomnd
	if p.Cut != "" {
		cuts := strings.SplitN(p.Cut, ":", 2)
		switch len(cuts) {
		case 1:
			p.cutFrom, p.cutTo = str.ParseInt(cuts[0]), 0
		case 2:
			p.cutFrom, p.cutTo = str.ParseInt(cuts[0]), str.ParseInt(cuts[1])
		}
	}

	return nil
}

// Config defines the config to capture a sub string from a string
type Config struct {
	Matches  []string `pflag:"前置匹配(子串包含)"`
	Splitter string   `plag:"切分分割符"`

	CaptureSplitSeq int    `pflag:"切分后取第几个子串(1开始)"`
	Capture         string `pflag:"匹配正则(优先级比锚点高)"`
	CaptureGroup    int    `pflag:"捕获组序号"`
	AnchorStart     string `pflag:"起始锚点(在capture为空时有效)"`
	AnchorEnd       string `pflag:"终止锚点(在capture为空时有效)"`
	CaptureCut      string `pflag:"切割，eg: 切除首尾字符 1:-1，切除尾部1一个字符:-1"`

	CmpRspSplitSeq     int    `pflag:"比较响应-切分后取第几个子串(1开始)"`
	CmpRspCapture      string `pflag:"比较响应-匹配正则(优先级比锚点高)"`
	CmpRspCaptureGroup int    `pflag:"比较响应-捕获组序号"`
	CmpRspAnchorStart  string `pflag:"比较响应-起始锚点(在capture为空时有效)"`
	CmpRspAnchorEnd    string `pflag:"比较响应-终止锚点(在capture为空时有效)"`
	CmpRspCut          string `pflag:"比较响应-切割，eg: 切除首尾字符 1:-1，切除尾部1一个字符:-1"`

	CmdRspOKLog  string `pflag:"比较响应-比较通过日志文件"`
	CmdRspBadLog string `pflag:"比较响应-比较失败日志文件"`

	capture       *DistractConfig
	cmpRspCapture *DistractConfig

	cmpRspOKLog  *log.Logger
	cmpRspBadLog *log.Logger
}

// Setup setup the Post p.
func (p *Config) Setup() error {
	p.capture = &DistractConfig{
		SplitSeq:     p.CaptureSplitSeq,
		Capture:      p.Capture,
		CaptureGroup: p.CaptureGroup,
		AnchorStart:  p.AnchorStart,
		AnchorEnd:    p.AnchorEnd,
		Cut:          p.CaptureCut,
	}

	if err := p.capture.setup(); err != nil {
		return err
	}

	return p.setupCmdResult()
}

func (p *Config) setupCmdResult() (err error) {
	p.cmpRspCapture = &DistractConfig{
		SplitSeq:     p.CmpRspSplitSeq,
		Capture:      p.CmpRspCapture,
		CaptureGroup: p.CmpRspCaptureGroup,
		AnchorStart:  p.CmpRspAnchorStart,
		AnchorEnd:    p.CmpRspAnchorEnd,
		Cut:          p.CmpRspCut,
	}

	if err := p.cmpRspCapture.setup(); err != nil {
		return err
	}

	if p.CmdRspOKLog != "" {
		p.cmpRspOKLog, err = createLog(p.CmdRspOKLog)
	}

	if p.CmdRspBadLog != "" {
		if p.CmdRspBadLog == p.CmdRspOKLog {
			p.cmpRspBadLog = p.cmpRspOKLog
		} else {
			p.cmpRspBadLog, err = createLog(p.CmdRspBadLog)
		}
	}

	if err == nil && p.cmpRspBadLog == nil {
		p.cmpRspBadLog = log.New(os.Stderr, "", log.LstdFlags)
	}

	return err
}

func createLog(filename string) (*log.Logger, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return log.New(f, "", log.LstdFlags), nil
}

func (p *Config) preMatches(s string) bool {
	for _, m := range p.Matches {
		if !strings.Contains(s, m) {
			return false
		}
	}

	return true
}

// CapturedStringResult defines the Captured result.
type CapturedStringResult struct {
	*Config
	Captured string
	parts    []string
}

// CaptureString captures string by config.
func (p *Config) CaptureString(s string) (*CapturedStringResult, error) {
	if !p.preMatches(s) {
		return nil, errors.New("s does not pass pre-matches")
	}

	parts := []string{s}

	if p.Splitter != "" {
		parts = strings.SplitN(s, p.Splitter, -1)
	}

	sub, err := p.capture.CaptureString(parts)
	if err != nil {
		return nil, err
	}

	return &CapturedStringResult{Config: p, Captured: sub, parts: parts}, err
}

// LogCmpRespOK logs the response comparing OK.
func (p *Config) LogCmpRespOK(url string, captured *CapturedStringResult, cmpResp string) {
	if p.cmpRspOKLog != nil {
		p.cmpRspOKLog.Printf("POST [%s] with [%s] response [%s] as expected",
			url, captured.Captured, cmpResp)
	}
}

// LogCmpRespBad logs the response comparing BAD.
func (p *Config) LogCmpRespBad(url string, captured *CapturedStringResult, cmpResp, realBody string) {
	p.cmpRspBadLog.Printf("POST [%s] with [%s] response [%s] different from expected [%s]",
		url, captured.Captured, cmpResp, realBody)
}

func (p *CapturedStringResult) IsCmdRespEmpty() bool { return p.cmpRspCapture.IsEmpty() }

func (p *CapturedStringResult) GetCmpResp() (string, error) {
	return p.cmpRspCapture.CaptureString(p.parts)
}

// CaptureString captures string by config.
func (p *DistractConfig) CaptureString(parts []string) (string, error) {
	s := parts[0]

	if p.SplitSeq > 0 {
		if p.SplitSeq > len(parts) {
			return "", fmt.Errorf("unable to get #%d of splitted parts", p.SplitSeq)
		}

		s = parts[p.SplitSeq-1]
	}

	if p.captureReg != nil {
		return p.captureByReg(s)
	}

	anchor, err := p.captureByAnchor(s)
	if err != nil {
		return "", err
	}

	return p.cut(anchor)
}

func (p *DistractConfig) captureByReg(line string) (string, error) {
	subs := p.captureReg.FindStringSubmatch(line)
	if len(subs) > p.CaptureGroup {
		return subs[p.CaptureGroup], nil
	}

	return "", errors.New("line does not match the Captured regular expression")
}

func (p *DistractConfig) captureByAnchor(line string) (string, error) {
	if p.AnchorStart != "" {
		pos := strings.Index(line, p.AnchorStart)
		if pos < 0 {
			return "", errors.New("line does not match anchor start")
		}

		line = line[pos+len(p.AnchorStart):]
	}

	if p.AnchorEnd != "" {
		pos := strings.Index(line, p.AnchorEnd)
		if pos < 0 {
			return "", errors.New("line does not match anchor end")
		}

		line = line[0:pos]
	}

	return line, nil
}

func (p *DistractConfig) cut(s string) (string, error) {
	if p.Cut == "" {
		return s, nil
	}

	if p.cutFrom < 0 {
		p.cutFrom += len(s)
	}

	if p.cutFrom < 0 {
		p.cutFrom = 0
	}

	if p.cutTo <= 0 {
		p.cutTo += len(s)
	}

	if p.cutTo > len(s) {
		p.cutTo = len(s)
	}

	return s[p.cutFrom:p.cutTo], nil
}
