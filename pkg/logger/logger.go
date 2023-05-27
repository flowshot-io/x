package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	zerologadapter "logur.dev/adapter/zerolog"
)

type (
	// Logger defines the interface for a logger.
	Logger interface {
		Trace(msg string, fields ...map[string]interface{})
		Debug(msg string, fields ...map[string]interface{})
		Info(msg string, fields ...map[string]interface{})
		Warn(msg string, fields ...map[string]interface{})
		Error(msg string, fields ...map[string]interface{})
	}

	// Options struct defines the logger options. Pretty determines whether logs will be pretty-printed.
	// LogLevel sets the level of logs to show (trace, debug, info, warn, error).
	Options struct {
		Pretty   bool
		LogLevel string
		Writer   io.Writer
	}

	// Option defines a function which sets an option on the Options struct.
	Option func(*Options)
)

// WithPretty sets the pretty flag on the Options struct.
func WithPretty() Option {
	return func(o *Options) {
		o.Pretty = true
	}
}

// WithLogLevel sets the log level on the Options struct.
func WithLogLevel(level string) Option {
	return func(o *Options) {
		o.LogLevel = level
	}
}

// WithWriter sets the writer on the Options struct.
func WithWriter(writer io.Writer) Option {
	return func(o *Options) {
		o.Writer = writer
	}
}

// New creates a new Logger based on provided Options.
func New(opts ...Option) Logger {
	options := &Options{
		Writer: os.Stderr,
	}

	for _, opt := range opts {
		opt(options)
	}

	zerolog.SetGlobalLevel(parseLogLevel(options.LogLevel))
	return zerologadapter.New(zerolog.New(getWriter(options)).With().Timestamp().Logger())
}

// NoOp returns a no-operation Logger which doesn't perform any logging operations.
func NoOp() Logger {
	return zerologadapter.New(zerolog.Nop())
}

// parseLogLevel parses the log level string and returns the corresponding zerolog.Level.
func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// getWriter creates and returns a writer based on the pretty flag. If pretty is true, it returns a ConsoleWriter.
// Otherwise, it returns the writer provided in options.
func getWriter(opts *Options) io.Writer {
	if opts.Pretty {
		return zerolog.ConsoleWriter{Out: opts.Writer, TimeFormat: zerolog.TimeFieldFormat}
	}

	return opts.Writer
}
