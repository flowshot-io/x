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

	Options struct {
		Pretty bool
	}
)

func New(opts *Options) Logger {
	if opts == nil {
		opts = &Options{}
	}

	var writer io.Writer
	if opts.Pretty {
		writer = zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		writer = os.Stderr
	}

	return zerologadapter.New(zerolog.New(writer).With().Timestamp().Logger())
}

func NoOp() Logger {
	return zerologadapter.New(zerolog.Nop())
}
