package liner

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/bingoohuang/gonet"
	"github.com/influxdata/tail"
	"github.com/sirupsen/logrus"
)

// Post process line and then POST it out
type Post struct {
	Matches []string // 前置匹配（子串包含）
	PostURL string   // POST url

	Capture      string // 匹配正则，优先级高
	CaptureGroup int    // 捕获组序号

	// 在Capture为空时，使用锚点定位
	AnchorStart string // 起始锚点
	AnchorEnd   string // 终止锚点

	client     *http.Client
	u          *url.URL
	q          url.Values
	captureReg *regexp.Regexp
}

// Setup setup the Post p.
func (p *Post) Setup() error {
	if p.PostURL != "" {
		p.u, _ = url.Parse(p.PostURL)
		p.q, _ = url.ParseQuery(p.u.RawQuery)
		p.client = &http.Client{
			Timeout: 60 * time.Second,
		}
	}

	var err error
	if p.Capture != "" {
		p.captureReg, err = regexp.Compile(p.Capture)
		if err != nil {
			return fmt.Errorf("compile regex %s  error %w", p.Capture, err)
		}
	}

	return nil
}

// ProcessLine process a line string.
func (p Post) ProcessLine(tailer *tail.Tail, line string, firstLine bool) error {
	if !p.matches(line) {
		return nil
	}

	captured := p.capture(line)
	if captured == "" {
		return nil
	}

	if p.PostURL != "" {
		p.postLine(firstLine, tailer.Filename, captured, line)
	}

	return nil
}

func (p Post) postLine(firstLine bool, filename, captured, line string) {
	//p.q.Add("filename", filename)
	//p.q.Add("firstLine", fmt.Sprintf("%v", firstLine))
	p.u.RawQuery = p.q.Encode()

	contentType := "text/plain; charset=utf-8"
	firstByte := captured[0]

	if firstByte == '{' || firstByte == '[' {
		contentType = "application/json; charset=utf-8"
	}

	start := time.Now()
	resp, err := p.client.Post(p.u.String(), contentType, strings.NewReader(captured)) // nolint

	if err != nil {
		logrus.Warnf("post: %s for line: %s error %+v", captured, line, err)
		return
	}

	status := resp.Status
	respBody := gonet.ReadString(resp.Body)
	logrus.Infof("post: %s cost: %v status: %s response: %s for line: %s",
		captured, time.Since(start), status, respBody, line)
}

func (p Post) matches(line string) bool {
	for _, m := range p.Matches {
		if !strings.Contains(line, m) {
			return false
		}
	}

	return true
}

func (p Post) capture(line string) string {
	if p.captureReg != nil {
		submatch := p.captureReg.FindStringSubmatch(line)
		if len(submatch) > p.CaptureGroup {
			return submatch[p.CaptureGroup]
		}

		return ""
	}

	if p.AnchorStart != "" {
		pos := strings.Index(line, p.AnchorStart)
		if pos < 0 {
			return ""
		}

		line = line[pos+len(p.AnchorStart):]
	}

	if p.AnchorEnd != "" {
		pos := strings.Index(line, p.AnchorEnd)
		if pos < 0 {
			return ""
		}

		line = line[0:pos]
	}

	return line
}
