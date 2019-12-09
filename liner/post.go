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
	Matches []string `pflag:"前置匹配(子串包含)"`
	PostURL string   `pflag:"POST URL"`

	Capture      string `pflag:"匹配正则(优先级比锚点高)"`
	CaptureGroup int    `pflag:"捕获组序号"`

	AnchorStart string `pflag:"起始锚点(在capture为空时有效)"`
	AnchorEnd   string `pflag:"终止锚点(在capture为空时有效)"`

	client     *http.Client
	postURL    *url.URL
	urlQuery   url.Values
	captureReg *regexp.Regexp
}

// Setup setup the Post p.
func (p *Post) Setup() error {
	if p.PostURL != "" {
		p.postURL, _ = url.Parse(p.PostURL)
		p.urlQuery, _ = url.ParseQuery(p.postURL.RawQuery)
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

// CloneURLValues clones an url.Values
func CloneURLValues(v url.Values) url.Values {
	// copy from https://golang.org/src/net/http/clone.go
	// http.Header and url.Values have the same representation, so temporarily
	// treat it like http.Header, which does have a clone
	return url.Values(http.Header(v).Clone())
}

func (p Post) postLine(firstLine bool, filename, captured, line string) {
	q := CloneURLValues(p.urlQuery)
	q.Add("filename", filename)
	q.Add("firstLine", fmt.Sprintf("%v", firstLine))

	u := p.postURL
	u.RawQuery = q.Encode()

	contentType := DetectContentType(captured)
	start := time.Now()
	postURL := u.String()
	logrus.Infof("postURL %s", postURL)
	resp, err := p.client.Post(postURL, contentType, strings.NewReader(captured)) // nolint

	if err != nil {
		logrus.Warnf("post: %s for line: %s error %+v", captured, line, err)
		return
	}

	status := resp.Status
	respBody := gonet.ReadString(resp.Body)

	logrus.Infof("original line: %s", line)
	logrus.Infof("post: %s cost: %v status: %s response: %s", captured, time.Since(start), status, respBody)
}

// DetectContentType detects content-type of body.
func DetectContentType(body string) string {
	switch body[0] {
	case '{', '[':
		return "application/json; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
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
		subs := p.captureReg.FindStringSubmatch(line)
		if len(subs) > p.CaptureGroup {
			return subs[p.CaptureGroup]
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
