package capture

import (
	"regexp"
	"strings"

	"github.com/bingoohuang/gou/str"
	"github.com/sirupsen/logrus"
)

// Pipe defines the filter interface for string filtering.
type Pipe interface {
	// Pipe returns filtered string slice or error.
	Filter(strs []string) []string
}

// ParseFilters parses the filters expression to a filter object.
func ParseFilters(s string) Pipe {
	filterExprs := str.SplitN(s, "|", true, true)
	filters := make([]Pipe, 0, len(filterExprs))

	for _, f := range filterExprs {
		fs := str.Fields(f, 2)
		name := fs[0]
		value := ""

		if len(fs) > 1 {
			value = fs[1]
		}

		if filterFactory, ok := filtersMap[name]; ok {
			filter := filterFactory(value)
			if filter != nil {
				filters = append(filters, filter)
			}
		}
	}

	if len(filters) == 0 {
		return nil
	}

	return &Filters{Filters: filters}
}

type Filters struct {
	Filters []Pipe
}

func (s Filters) Filter(strs []string) []string {
	out := strs

	for _, f := range s.Filters {
		out = f.Filter(out)
	}

	return out
}

type FilterFactory func(expr string) Pipe

// nolint gochecknoinits
var filtersMap = make(map[string]FilterFactory)

type Contains struct {
	Subs []string
}

func (c Contains) Filter(strs []string) []string {
	for _, i := range strs {
		if c.matches(i) {
			return strs
		}
	}

	return []string{}
}

func (c Contains) matches(str string) bool {
	for _, sub := range c.Subs {
		if strings.Contains(str, sub) {
			return true
		}
	}

	return false
}

// nolint gochecknoinits
func init() {
	filtersMap["contains"] = func(expr string) Pipe {
		subs := make([]string, 0)

		for _, sub := range strings.Fields(expr) {
			subs = append(subs, sub)
		}

		if len(subs) == 0 {
			return nil
		}

		return &Contains{Subs: subs}
	}
}

type Grep struct {
	Exprs []*regexp.Regexp
}

// nolint gochecknoinits
func init() {
	filtersMap["grep"] = func(expr string) Pipe {
		exprs := make([]*regexp.Regexp, 0)

		for _, sub := range strings.Fields(expr) {
			reg, err := regexp.Compile(sub)
			if err != nil {
				logrus.Panicf("unable to compile regex %s error: %v", sub, err)
			}

			exprs = append(exprs, reg)
		}

		if len(exprs) == 0 {
			return nil
		}

		return &Grep{Exprs: exprs}
	}
}

func (s Grep) Filter(strs []string) []string {
	for _, i := range strs {
		if s.matches(i) {
			return strs
		}
	}

	return []string{}
}

func (s Grep) matches(str string) bool {
	for _, i := range s.Exprs {
		if i.FindString(str) != "" {
			return true
		}
	}

	return false
}

type Reg struct {
	Expr   *regexp.Regexp
	Groups []int
}

// nolint gochecknoinits
func init() {
	filtersMap["reg"] = func(expr string) Pipe {
		fields := strings.Fields(expr)
		if len(fields) == 0 {
			return nil
		}

		r, err := regexp.Compile(fields[0])
		if err != nil {
			logrus.Panicf("unable to compile regex %s error: %v", fields[0], err)
		}

		groups := make([]int, 0)

		if len(fields) == 1 {
			groups = append(groups, 0)
		} else {
			for i := 1; i < len(fields); i++ {
				groups = append(groups, str.ParseInt(fields[i]))
			}
		}

		return &Reg{Expr: r, Groups: groups}
	}
}

func (s Reg) Filter(strs []string) []string {
	out := make([]string, 0)

	for _, i := range strs {
		out = append(out, s.capture(i)...)
	}

	return out
}

func (s Reg) capture(str string) []string {
	subs := s.Expr.FindStringSubmatch(str)
	outs := make([]string, 0, len(subs))

	for _, i := range s.Groups {
		if i >= 0 && i < len(subs) {
			outs = append(outs, subs[i])
		}
	}

	return outs
}

type Trim struct{}

// nolint gochecknoinits
func init() {
	filtersMap["trim"] = func(expr string) Pipe {
		return &Trim{}
	}
}

func (s Trim) Filter(strs []string) []string {
	out := make([]string, 0)

	for _, i := range strs {
		if ii := strings.TrimSpace(i); ii != "" {
			out = append(out, ii)
		}
	}

	return out
}

type Anchor struct {
	Start string
	End   string

	IncludeStart bool
	IncludeEnd   bool
}

// nolint gochecknoinits
func init() {
	filtersMap["anchor"] = func(expr string) Pipe {
		props := parseProps(strings.Fields(expr))

		return &Anchor{
			Start:        props["start"],
			End:          props["end"],
			IncludeStart: parseBool(props["includeStart"]),
			IncludeEnd:   parseBool(props["includeEnd"]),
		}
	}
}

func parseBool(s string) bool {
	switch strings.ToLower(s) {
	case "true", "yes", "on", "1":
		return true
	}

	return false
}

func parseProps(fields []string) map[string]string {
	props := make(map[string]string)

	for i := 0; i < len(fields); i++ {
		k, v := str.Split2(fields[i], "=", true, true)
		props[k] = v
	}

	return props
}

func (s Anchor) Filter(strs []string) []string {
	out := make([]string, 0)

	for _, i := range strs {
		if ii := s.capture(i); ii != "" {
			out = append(out, ii)
		}
	}

	return out
}

func (s Anchor) capture(str string) string {
	if s.Start != "" {
		pos := strings.Index(str, s.Start)
		if pos < 0 {
			return ""
		}

		if s.IncludeStart {
			str = str[pos:]
		} else {
			str = str[pos+len(s.Start):]
		}
	}

	if s.End != "" {
		pos := strings.LastIndex(str, s.End)
		if pos < 0 {
			return ""
		}

		if s.IncludeEnd {
			str = str[0 : pos+len(s.End)]
		} else {
			str = str[0:pos]
		}
	}

	return str
}

type Join struct {
	By string
}

func (j Join) Filter(strs []string) []string {
	return []string{strings.Join(strs, j.By)}
}

// nolint gochecknoinits
func init() {
	filtersMap["join"] = func(expr string) Pipe {
		props := parseProps(strings.Fields(expr))

		return &Join{By: props["by"]}
	}
}

type Split struct {
	By    string
	N     string
	Keeps []int
}

// nolint gochecknoinits
func init() {
	filtersMap["split"] = func(expr string) Pipe {
		props := parseProps(strings.Fields(expr))

		return &Split{
			By:    props["by"],
			N:     props["n"],
			Keeps: parseInts(props["keeps"]),
		}
	}
}

func parseInts(s string) []int {
	out := make([]int, 0)

	for _, v := range str.SplitN(s, ",", true, true) {
		out = append(out, str.ParseInt(v))
	}

	return out
}

func (s Split) Filter(strs []string) []string {
	out := make([]string, 0)

	if s.By == "" {
		for _, i := range strs {
			out = append(out, strings.Fields(i)...)
		}
	} else {
		n := -1

		if s.N != "" {
			n = str.ParseInt(s.N)
		}

		for _, i := range strs {
			out = append(out, strings.SplitN(i, s.By, n)...)
		}
	}

	kept := make([]string, 0, len(out))

	for _, i := range s.Keeps {
		if i >= 0 && i < len(out) {
			kept = append(kept, out[i])
		}
	}

	return kept
}

type Cutter struct {
	Start string
	End   string
}

// nolint gomnd
func init() {
	filtersMap["cut"] = func(expr string) Pipe {
		fs := strings.SplitN(expr, ":", 2)
		start, end := "", ""

		switch len(fs) {
		case 2:
			start, end = fs[0], fs[1]
		case 1:
			start = fs[0]
		}

		return &Cutter{Start: start, End: end}
	}
}

func (s Cutter) Filter(strs []string) []string {
	out := make([]string, 0)

	startIndex := 0
	if s.Start != "" {
		startIndex = str.ParseInt(s.Start)
	}

	endIndex := -1
	if s.End != "" {
		endIndex = str.ParseInt(s.End)
	}

	for _, i := range strs {
		out = append(out, cut(i, startIndex, endIndex))
	}

	return out
}

func cut(s string, start, end int) string {
	if start < 0 {
		start += len(s)
	}

	if start < 0 {
		start = 0
	}

	if end < 0 {
		end += len(s)
	}

	if end > len(s) {
		end = len(s)
	}

	return s[start:end]
}
