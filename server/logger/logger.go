package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Logger struct {
	threshold Level // values above threshold will be outputted
	output    io.Writer
	date      bool
}

func New(threshold Level, opts ...Options) *Logger {
	lgs := &Logger{threshold: threshold, output: os.Stdout}

	for _, f := range opts {
		f(lgs)
	}
	return lgs
}

func (lg *Logger) Infof(format string, args ...any) {
	format += "\n"
	if lg.threshold > Info {
		return
	}
	prefix := TemplateInfo
	if lg.date {
		prefix += TemplateSeparator + time.Now().String()
	}
	lg.logf(format, args...)
}

func (lg *Logger) WARNf(format string, args ...any) {
	format += "\n"
	if lg.threshold > Warn {
		return
	}
	prefix := TemplateWarn
	if lg.date {
		prefix += TemplateSeparator + time.Now().String()
	}
	lg.logf(format, args...)
}

func (lg *Logger) Errorf(format string, args ...any) {
	format += "\n"
	if lg.threshold > Error {
		return
	}
	prefix := TemplateError
	if lg.date {
		prefix += TemplateSeparator + time.Now().String()
	}
	lg.logf(format, args...)
}

func (lg *Logger) Fatalf(format string, args ...any) {
	if lg.threshold > Fatal {
		return
	}
	prefix := TemplateFatal
	if lg.date {
		prefix += TemplateSeparator + time.Now().String()
	}
	lg.logf(format, args...)
	os.Exit(1)
}

func (lg *Logger) logf(format string, args ...any) {
	_, _ = fmt.Fprintf(lg.output, format, args...)
}
