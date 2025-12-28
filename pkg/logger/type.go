package logger

import "io"

// LoggerType represents different logger implementations
type LoggerType string

const (
	LoggerTypeZerolog LoggerType = "zerolog"
)

// LogLevel represents logging level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// LogFormat represents log output format
type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

type LoggerConfig struct {
	Type   LoggerType
	Level  LogLevel
	Output io.Writer
	Format LogFormat
}

// FieldType represents the type of a log field
type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeInt     FieldType = "int"
	FieldTypeInt64   FieldType = "int64"
	FieldTypeUint    FieldType = "uint"
	FieldTypeFloat64 FieldType = "float64"
	FieldTypeBool    FieldType = "bool"
	FieldTypeError   FieldType = "error"
	FieldTypeAny     FieldType = "any"
)

type LogField struct {
	Key   string
	Value interface{}
	Type  FieldType
}

// LoggerAdapter is the interface that different logger implementations must satisfy
type LoggerAdapter interface {
	// Log at different levels with message and fields
	Log(level LogLevel, msg string, fields []LogField)

	// Formatted logging
	Logf(level LogLevel, template string, args ...interface{})
}
