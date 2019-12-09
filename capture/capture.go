package capture

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Config defines the config to capture a sub string from a string
type Config struct {
	Matches      []string `pflag:"前置匹配(子串包含)"`
	Capture      string   `pflag:"匹配正则(优先级比锚点高)"`
	CaptureGroup int      `pflag:"捕获组序号"`

	AnchorStart string `pflag:"起始锚点(在capture为空时有效)"`
	AnchorEnd   string `pflag:"终止锚点(在capture为空时有效)"`

	captureReg *regexp.Regexp
}

// Setup setup the Post p.
func (p *Config) Setup() error {
	var err error
	if p.Capture != "" {
		p.captureReg, err = regexp.Compile(p.Capture)
		if err != nil {
			return fmt.Errorf("compile regex %s  error %w", p.Capture, err)
		}
	}

	return nil
}

func (p Config) preMatches(s string) bool {
	for _, m := range p.Matches {
		if !strings.Contains(s, m) {
			return false
		}
	}

	return true
}

// CaptureString captures string by config.
func (p Config) CaptureString(s string) (string, error) {
	if !p.preMatches(s) {
		return "", errors.New("s does not pass pre-matches")
	}

	if p.captureReg != nil {
		return p.captureByReg(s)
	}

	return p.captureByAnchor(s)
}

func (p Config) captureByReg(line string) (string, error) {
	subs := p.captureReg.FindStringSubmatch(line)
	if len(subs) > p.CaptureGroup {
		return subs[p.CaptureGroup], nil
	}

	return "", errors.New("line does not match the Capture regular expression")
}

func (p Config) captureByAnchor(line string) (string, error) {
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
