package logger

import (
	"os"
	"sync"
)

var (
	loggerInstance Logger
	globalOnce     sync.Once
)

// Logger is the main logging interface with builder pattern
type Logger interface {
	// Create log entries with different levels (builder pattern)
	Debug(msg string) Logger
	Info(msg string) Logger
	Warn(msg string) Logger
	Error(msg string) Logger
	Fatal(msg string) Logger

	// Builder methods - return Logger for chaining
	WithAny(key string, value interface{}) Logger
	WithString(key, value string) Logger
	WithInt(key string, value int) Logger
	WithInt64(key string, value int64) Logger
	WithUint(key string, value uint) Logger
	WithFloat64(key string, value float64) Logger
	WithBool(key string, value bool) Logger
	WithError(err error) Logger
	WithFields(fields map[string]interface{}) Logger

	// Log method - writes the accumulated log entry
	Log()

	// Formatted logging methods
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
}

func defaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Type:   LoggerTypeZerolog,
		Level:  LogLevelInfo,
		Output: os.Stdout,
		Format: LogFormatJSON,
	}
}

func GetLogger() Logger {
	var err error

	if loggerInstance == nil {
		globalOnce.Do(func() {
			loggerInstance, err = newLogger(defaultLoggerConfig())
			if err != nil {
				panic(err)
			}
		})
	}
	return loggerInstance
}

func SetLogger(logConfig LoggerConfig) {
	var err error
	loggerInstance, err = newLogger(logConfig)
	if err != nil {
		panic(err)
	}
}

func Debug(msg string) Logger {
	return GetLogger().Debug(msg)
}

func Info(msg string) Logger {
	return GetLogger().Info(msg)
}

func Warn(msg string) Logger {
	return GetLogger().Warn(msg)
}

func Error(msg string) Logger {
	return GetLogger().Error(msg)
}

func Fatal(msg string) Logger {
	return GetLogger().Fatal(msg)
}
