package capture

import (
	"log"
	"os"
)

// Config defines the config to capture a sub string from a string
type Config struct {
	Capture   string `pflag:"捕获表达式"`
	ExpectRsp string `pflag:"期待响应表达式"`

	RspOKLog   string `pflag:"比较响应-比较通过日志文件名"`
	RspFailLog string `pflag:"比较响应-比较失败日志文件名"`

	cmpRspOKLog  *log.Logger
	cmpRspBadLog *log.Logger

	captureFilter   Pipe
	expectRspFilter Pipe
}

// Setup setup the Post p.
func (p *Config) Setup() error {
	p.captureFilter = ParseFilters(p.Capture)
	p.expectRspFilter = ParseFilters(p.ExpectRsp)

	return p.setupCmdResult()
}

func (p *Config) setupCmdResult() (err error) {
	if p.RspOKLog != "" {
		p.cmpRspOKLog, err = createLog(p.RspOKLog)
	}

	if p.RspFailLog != "" {
		if p.RspFailLog == p.RspOKLog {
			p.cmpRspBadLog = p.cmpRspOKLog
		} else {
			p.cmpRspBadLog, err = createLog(p.RspFailLog)
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

// CapturedStringResult defines the Captured result.
type CapturedStringResult struct {
	*Config
	Captured string
}

// CaptureString captures string by config.
func (p *Config) CaptureString(s string) *CapturedStringResult {
	outs := p.captureFilter.Filter([]string{s})
	if len(outs) == 0 {
		return nil
	}

	return &CapturedStringResult{Config: p, Captured: outs[0]}
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

func (p *CapturedStringResult) IsCmdRespEmpty() bool { return p.expectRspFilter == nil }

func (p *CapturedStringResult) GetCmpResp(s string) string {
	out := p.expectRspFilter.Filter([]string{s})
	if len(out) > 0 {
		return out[0]
	}

	return ""
}
