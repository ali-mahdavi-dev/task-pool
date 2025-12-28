package logger

import "fmt"

// logger is the implementation of Logger interface
type logger struct {
	adapter LoggerAdapter
	config  LoggerConfig
	fields  []LogField // accumulated fields from builder
	level   LogLevel   // log level for builder
	msg     string     // message for builder
}

// newLogger creates a new logger instance
func newLogger(config LoggerConfig) (Logger, error) {
	var adapter LoggerAdapter
	var err error

	switch config.Type {
	case LoggerTypeZerolog:
		adapter, err = newZerologAdapter(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create zerolog adapter: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported logger type: %s", config.Type)
	}

	return &logger{
		adapter: adapter,
		config:  config,
		fields:  make([]LogField, 0),
	}, nil
}

func (l *logger) Debug(msg string) Logger {
	l.level = LogLevelDebug
	l.msg = msg
	l.fields = make([]LogField, 0)
	return l
}

func (l *logger) Info(msg string) Logger {
	l.level = LogLevelInfo
	l.msg = msg
	l.fields = make([]LogField, 0)
	return l
}

func (l *logger) Warn(msg string) Logger {
	l.level = LogLevelWarn
	l.msg = msg
	l.fields = make([]LogField, 0)
	return l
}

func (l *logger) Error(msg string) Logger {
	l.level = LogLevelError
	l.msg = msg
	l.fields = make([]LogField, 0)
	return l
}

func (l *logger) Fatal(msg string) Logger {
	l.level = LogLevelFatal
	l.msg = msg
	l.fields = make([]LogField, 0)
	return l
}

func (l *logger) Debugf(template string, args ...interface{}) {
	if l.shouldLog(LogLevelDebug) {
		l.adapter.Logf(LogLevelDebug, template, args...)
	}
}

func (l *logger) Infof(template string, args ...interface{}) {
	if l.shouldLog(LogLevelInfo) {
		l.adapter.Logf(LogLevelInfo, template, args...)
	}
}

func (l *logger) Warnf(template string, args ...interface{}) {
	if l.shouldLog(LogLevelWarn) {
		l.adapter.Logf(LogLevelWarn, template, args...)
	}
}

func (l *logger) Errorf(template string, args ...interface{}) {
	if l.shouldLog(LogLevelError) {
		l.adapter.Logf(LogLevelError, template, args...)
	}
}

func (l *logger) Fatalf(template string, args ...interface{}) {
	l.adapter.Logf(LogLevelFatal, template, args...)
}

// Logger builder methods - mutate self and return self

func (l *logger) WithAny(key string, value interface{}) Logger {
	l.fields = append(l.fields, LogField{
		Key:   key,
		Value: value,
		Type:  FieldTypeAny,
	})
	return l
}

func (l *logger) WithString(key, value string) Logger {
	l.fields = append(l.fields, LogField{
		Key:   key,
		Value: value,
		Type:  FieldTypeString,
	})
	return l
}

func (l *logger) WithInt(key string, value int) Logger {
	l.fields = append(l.fields, LogField{
		Key:   key,
		Value: value,
		Type:  FieldTypeInt,
	})
	return l
}

func (l *logger) WithInt64(key string, value int64) Logger {
	l.fields = append(l.fields, LogField{
		Key:   key,
		Value: value,
		Type:  FieldTypeInt64,
	})
	return l
}

func (l *logger) WithUint(key string, value uint) Logger {
	l.fields = append(l.fields, LogField{
		Key:   key,
		Value: value,
		Type:  FieldTypeUint,
	})
	return l
}

func (l *logger) WithFloat64(key string, value float64) Logger {
	l.fields = append(l.fields, LogField{
		Key:   key,
		Value: value,
		Type:  FieldTypeFloat64,
	})
	return l
}

func (l *logger) WithBool(key string, value bool) Logger {
	l.fields = append(l.fields, LogField{
		Key:   key,
		Value: value,
		Type:  FieldTypeBool,
	})
	return l
}

func (l *logger) WithError(err error) Logger {
	if err != nil {
		l.fields = append(l.fields, LogField{
			Key:   "error",
			Value: err,
			Type:  FieldTypeError,
		})
	}
	return l
}

func (l *logger) WithFields(fields map[string]interface{}) Logger {
	for k, v := range fields {
		l.fields = append(l.fields, LogField{
			Key:   k,
			Value: v,
			Type:  FieldTypeAny,
		})
	}
	return l
}

func (l *logger) Log() {
	if l.msg == "" {
		return // No message set, can't log
	}

	if !l.shouldLog(l.level) {
		return
	}

	l.adapter.Log(l.level, l.msg, l.fields)

	// Clear fields after logging
	l.fields = make([]LogField, 0)
	l.msg = ""
	l.level = ""
}

// shouldLog checks if the message should be logged based on log level
func (l *logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}

	return levels[level] >= levels[l.config.Level]
}
