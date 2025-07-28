package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// Config represents logger configuration
type Config struct {
	// Output is the output destination
	Output io.Writer
	// Level is the minimum level to log
	Level zerolog.Level
	// Pretty enables pretty logging (human-readable)
	Pretty bool
	// WithCaller adds caller info to log
	WithCaller bool
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() Config {
	return Config{
		Output:     os.Stdout,
		Level:      zerolog.InfoLevel,
		Pretty:     false,
		WithCaller: true,
	}
}

// Init initializes the global logger with the given configuration
func Init(config Config) {
	// Set global logger
	zerolog.SetGlobalLevel(config.Level)
	
	// Configure logger output
	var output io.Writer = config.Output
	if config.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        config.Output,
			TimeFormat: time.RFC3339,
		}
	}
	
	// Configure logger context
	logger := zerolog.New(output).With().Timestamp()
	if config.WithCaller {
		logger = logger.Caller()
	}
	
	// Set global logger
	log.Logger = logger.Logger()
}

// Debug logs a debug message
func Debug(msg string, fields ...Field) {
	event := log.Debug()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Info logs an info message
func Info(msg string, fields ...Field) {
	event := log.Info()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Warn logs a warning message
func Warn(msg string, fields ...Field) {
	event := log.Warn()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Error logs an error message
func Error(msg string, err error, fields ...Field) {
	event := log.Error()
	if err != nil {
		event = event.Err(err)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, err error, fields ...Field) {
	event := log.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// GetLoggerFromContext retrieves the logger from the gin context
func GetLoggerFromContext(c *gin.Context) *zerolog.Logger {
	loggerInterface, exists := c.Get("logger")
	if !exists {
		return &log.Logger
	}
	
	logger, ok := loggerInterface.(*zerolog.Logger)
	if !ok {
		return &log.Logger
	}
	
	return logger
}

// Field creators
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

func Str(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Uint(key string, value uint) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

// SetLevel sets the global log level
func SetLevel(level string) error {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		return fmt.Errorf("unknown log level: %s", level)
	}
	return nil
}