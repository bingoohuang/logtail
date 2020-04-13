package liner

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bingoohuang/logtail/capture"

	"github.com/bingoohuang/gonet"
	"github.com/influxdata/tail"
	"github.com/sirupsen/logrus"
)

// Post process line and then POST it out
type Post struct {
	PostURL string `pflag:"POST URL"`

	capture.Config

	client   *http.Client
	postURL  *url.URL
	urlQuery url.Values
}

// Setup setup the Post p.
func (p *Post) Setup() error {
	if p.PostURL != "" {
		p.postURL, _ = url.Parse(p.PostURL)
		p.urlQuery, _ = url.ParseQuery(p.postURL.RawQuery)
		p.client = &http.Client{
			Timeout: 60 * time.Second, // nolint gomnd
		}
	} else {
		logrus.Warnf("PostURL is blank")
	}

	return p.Config.Setup()
}

// ProcessLine process a line string.
func (p Post) ProcessLine(tailer *tail.Tail, line string, firstLine bool) error {
	captured, err := p.CaptureString(line)
	if err != nil || captured == "" {
		logrus.Debugf("no capture for line %s error %v", line, err)
		return nil
	}

	logrus.Infof("captured %s from line %s", captured, line)

	if p.PostURL != "" {
		p.postLine(captured)
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

func (p Post) postLine(line string) {
	q := CloneURLValues(p.urlQuery)
	//q.Add("filename", filename)
	//q.Add("firstLine", fmt.Sprintf("%v", firstLine))

	u := p.postURL
	u.RawQuery = q.Encode()

	contentType := DetectContentType(line)
	start := time.Now()
	postURL := u.String()
	logrus.Infof("postURL %s", postURL)
	resp, err := p.client.Post(postURL, contentType, strings.NewReader(line)) // nolint

	if err != nil {
		logrus.Warnf("post error %+v", err)
		return
	}

	status := resp.Status
	respBody := gonet.ReadString(resp.Body)

	logrus.Infof("post cost: %v status: %s response: %s", time.Since(start), status, respBody)
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
