package logger

type Level byte

const (
	Info Level = iota
	Warn
	Error
	Fatal
)

const (
	TemplateInfo      = "< Info > "
	TemplateWarn      = "< WARN > "
	TemplateError     = "< Error > "
	TemplateFatal     = "< FATAL>"
	TemplateSeparator = "::"
)
