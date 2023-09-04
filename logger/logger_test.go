package logger

import (
	"testing"
)

const (
	inf    = "some message1"
	wrn    = "some other message"
	errorm = "error message here"
)

type twriter struct {
	contents string
}

func TestLogger(t *testing.T) {
	type testCase struct {
		level Level
		want  string
	}
	tt := map[string]testCase{
		"info": {
			level: Info,
			want:  TemplateInfo + inf + "\n" + TemplateWarn + wrn + "\n" + TemplateError + errorm + "\n",
		},
		"warn": {
			level: Warn,
			want:  TemplateWarn + wrn + "\n" + TemplateError + errorm + "\n",
		},
		"error": {
			level: Error,
			want:  TemplateError + errorm + "\n",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			tw := &twriter{}
			logger := New(tc.level, ToOutput(tw))
			logger.Infof(inf)
			logger.Warnf(wrn)
			logger.Errorf(errorm)
			got := tw.contents
			if tc.want != got {
				t.Errorf("wanted %v, but got %v", tc.want, got)
			}

		})
	}
}

func (tw *twriter) Write(p []byte) (n int, err error) {
	tw.contents += string(p)
	return len(p), nil
}
