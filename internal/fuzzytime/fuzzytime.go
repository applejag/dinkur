package fuzzytime

import (
	"errors"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

var (
	ErrUnknownFormat = errors.New("unknown time format")
)

var w *when.Parser

func init() {
	w = when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)
}

func Parse(s string) (time.Time, error) {
	if strings.EqualFold(s, "now") {
		return time.Now(), nil
	}
	if t, ok := ParseDelta(s); ok {
		return t, nil
	}
	return ParseWhen(s)
}

func ParseWhen(s string) (time.Time, error) {
	r, err := w.Parse(s, time.Now())
	if err != nil {
		return time.Time{}, err
	}
	if r == nil {
		return time.Time{}, ErrUnknownFormat
	}
	return r.Time, nil
}

func ParseDelta(s string) (time.Time, bool) {
	if len(s) < 3 {
		return time.Time{}, false
	}
	if s[0] != '+' && s[0] != '-' {
		return time.Time{}, false
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return time.Time{}, false
	}
	return time.Now().Add(d), true
}
