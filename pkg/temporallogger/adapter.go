package temporallogger

import "github.com/flowshot-io/x/pkg/logger"

type TemporalLoggerAdapter struct {
	logger logger.Logger
}

func New(logger logger.Logger) *TemporalLoggerAdapter {
	return &TemporalLoggerAdapter{
		logger: logger,
	}
}

func (t *TemporalLoggerAdapter) Debug(msg string, keyvals ...interface{}) {
	t.logger.Debug(msg, keyvalsToFields(keyvals...))
}

func (t *TemporalLoggerAdapter) Info(msg string, keyvals ...interface{}) {
	t.logger.Info(msg, keyvalsToFields(keyvals...))
}

func (t *TemporalLoggerAdapter) Warn(msg string, keyvals ...interface{}) {
	t.logger.Warn(msg, keyvalsToFields(keyvals...))
}

func (t *TemporalLoggerAdapter) Error(msg string, keyvals ...interface{}) {
	t.logger.Error(msg, keyvalsToFields(keyvals...))
}

// Convert keyvals (slice of alternating keys and values) to a map[string]interface{}
func keyvalsToFields(keyvals ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue // or you can handle it differently
		}

		val := keyvals[i+1]
		fields[key] = val
	}

	return fields
}
