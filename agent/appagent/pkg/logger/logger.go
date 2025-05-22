package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

// Log is the global logger instance
var Log zerolog.Logger

// Config holds the configuration for the logger
type Config struct {
	Level      string `mapstructure:"level"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// Init initializes the logger with the given configuration
func Init(cfg *Config) error {
	// Create logs directory if it doesn't exist
	if cfg.Output == "file" || cfg.Output == "both" {
		logDir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Set up the logger
	var writers []io.Writer

	// Configure console output
	if cfg.Output == "console" || cfg.Output == "both" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		writers = append(writers, consoleWriter)
	}

	// Configure file output
	if cfg.Output == "file" || cfg.Output == "both" {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,    // megabytes
			MaxBackups: cfg.MaxBackups, // number of backups
			MaxAge:     cfg.MaxAge,     // days
			Compress:   cfg.Compress,   // compress rotated files
		}
		writers = append(writers, fileWriter)
	}

	// Create multi-writer if needed
	var w io.Writer
	if len(writers) > 1 {
		w = io.MultiWriter(writers...)
	} else if len(writers) == 1 {
		w = writers[0]
	} else {
		// Default to console if no output is specified
		w = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Set global time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Set log level
	level, err := parseLogLevel(cfg.Level)
	if err != nil {
		return err
	}

	// Create logger
	Log = zerolog.New(w).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	return nil
}

// parseLogLevel converts a string log level to zerolog.Level
func parseLogLevel(level string) (zerolog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// Debug logs a debug message
func Debug() *zerolog.Event {
	return Log.Debug()
}

// Info logs an info message
func Info() *zerolog.Event {
	return Log.Info()
}

// Warn logs a warning message
func Warn() *zerolog.Event {
	return Log.Warn()
}

// Error logs an error message
func Error() *zerolog.Event {
	return Log.Error()
}

// Fatal logs a fatal message and exits
func Fatal() *zerolog.Event {
	return Log.Fatal()
}

// Panic logs a panic message and panics
func Panic() *zerolog.Event {
	return Log.Panic()
}
