package logger

import "io"

type Options func(*Logger)

func ToOutput(output io.Writer) Options {
	return func(lg *Logger) {
		lg.output = output
	}
}

func DateOpts(output bool) Options {
	return func(lg *Logger) {
		lg.date = output
	}
}
