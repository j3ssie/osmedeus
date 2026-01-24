package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	mu           sync.Mutex
)

// Config holds logger configuration
type Config struct {
	Level       string // debug, info, warn, error
	Development bool   // use development mode with more verbose output
	Verbose     bool   // show source file/caller information
	Silent      bool   // silent mode - disable console output
	OutputPaths []string
	LogDir      string // directory for log files
	LogFile     string // direct path to log file (takes precedence over LogDir)
}

// ColoredTimeEncoder formats time as ISO 8601 with grey color
func ColoredTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	const grey = "\033[90m"
	const reset = "\033[0m"
	timestamp := t.Format("2006-01-02T15:04:05-07:00")
	enc.AppendString(grey + timestamp + reset)
}

// ColoredLevelEncoder formats level as bold + colored
func ColoredLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	const (
		reset   = "\033[0m"
		bold    = "\033[1m"
		red     = "\033[31m"
		yellow  = "\033[33m"
		magenta = "\033[35m"
		cyan    = "\033[36m"
	)

	var color string
	switch l {
	case zapcore.DebugLevel:
		color = magenta
	case zapcore.InfoLevel:
		color = cyan
	case zapcore.WarnLevel:
		color = yellow
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		color = red
	default:
		color = ""
	}

	enc.AppendString(bold + color + l.CapitalString() + reset)
}

// PlainTimeEncoder formats time as ISO 8601 without color (for file logging)
func PlainTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05-07:00"))
}

// ColoredCallerEncoder formats caller location in bright cyan color
func ColoredCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	const brightCyan = "\033[96m"
	const reset = "\033[0m"
	enc.AppendString(brightCyan + caller.TrimmedPath() + reset)
}

// coloredConsoleEncoder wraps a console encoder to color JSON fields in gray
type coloredConsoleEncoder struct {
	zapcore.Encoder
	pool buffer.Pool
}

// newColoredConsoleEncoder creates a new colored console encoder
func newColoredConsoleEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return &coloredConsoleEncoder{
		Encoder: zapcore.NewConsoleEncoder(cfg),
		pool:    buffer.NewPool(),
	}
}

// Clone implements zapcore.Encoder
func (e *coloredConsoleEncoder) Clone() zapcore.Encoder {
	return &coloredConsoleEncoder{
		Encoder: e.Encoder.Clone(),
		pool:    e.pool,
	}
}

// EncodeEntry implements zapcore.Encoder with JSON field coloring
func (e *coloredConsoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return buf, err
	}

	// Post-process to color JSON-like content in gray
	content := buf.String()
	colored := colorJSONFields(content)

	newBuf := e.pool.Get()
	newBuf.AppendString(colored)
	buf.Free()

	return newBuf, nil
}

// ansiPattern matches ANSI escape sequences
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape sequences from a string
func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

// colorJSONFields wraps JSON-like content {...} in gray color.
// It also strips any ANSI codes from within the JSON block to prevent
// escape sequences from appearing in structured log fields.
func colorJSONFields(s string) string {
	const gray = "\033[90m"
	const reset = "\033[0m"

	// Match JSON-like patterns: {...}
	re := regexp.MustCompile(`(\{[^}]+\})`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Strip any ANSI codes from within the JSON block first
		clean := stripANSI(match)
		return gray + clean + reset
	})
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:       "info",
		Development: false,
		OutputPaths: []string{"stdout"},
	}
}

// Init initializes the global logger with the given configuration
// This can be called multiple times to reconfigure the logger
func Init(cfg Config) error {
	mu.Lock()
	defer mu.Unlock()

	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// NewLogger creates a new zap logger with the given configuration
func NewLogger(cfg Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Console encoder config with colored output
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:          "timestamp",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "", // Set conditionally below
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "message",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      ColoredLevelEncoder, // Bold + colored level
		EncodeTime:       PlainTimeEncoder,    // Plain ISO 8601 timestamp (avoid ANSI codes in data fields)
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     ColoredCallerEncoder, // Bright cyan caller location
		ConsoleSeparator: " ",                  // Single space between fields
	}

	// File encoder config without colors
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // Plain uppercase level
		EncodeTime:     PlainTimeEncoder,            // Plain ISO 8601 timestamp
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Only show caller in verbose mode
	if cfg.Verbose {
		consoleEncoderConfig.CallerKey = "caller"
		fileEncoderConfig.CallerKey = "caller"
	}

	var cores []zapcore.Core

	// Console output with colors (skip if silent mode)
	if !cfg.Silent {
		consoleEncoder := newColoredConsoleEncoder(consoleEncoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// File output if LogFile or LogDir is specified (without colors)
	var logFilePath string
	if cfg.LogFile != "" {
		// Direct log file path takes precedence
		logFilePath = cfg.LogFile
		// Ensure parent directory exists
		if dir := filepath.Dir(logFilePath); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
		}
	} else if cfg.LogDir != "" {
		// Fall back to LogDir/osmedeus.log
		if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
			return nil, err
		}
		logFilePath = filepath.Join(cfg.LogDir, "osmedeus.log")
	}

	if logFilePath != "" {
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		jsonEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
		fileCore := zapcore.NewCore(
			jsonEncoder,
			zapcore.AddSync(file),
			level,
		)
		cores = append(cores, fileCore)

		// Print DEBUG message showing log file destination
		if !cfg.Silent {
			const magenta = "\033[35m"
			const bold = "\033[1m"
			const gray = "\033[90m"
			const reset = "\033[0m"
			timestamp := time.Now().Format("2006-01-02T15:04:05-07:00")
			fmt.Printf("%s%s %s%sDEBUG%s Logging to file: %s\n",
				gray, timestamp, reset, bold+magenta, reset, logFilePath)
		}
	}

	// If no cores (silent mode with no file output), use a nop core
	if len(cores) == 0 {
		cores = append(cores, zapcore.NewNopCore())
	}

	core := zapcore.NewTee(cores...)

	var opts []zap.Option
	if cfg.Development {
		opts = append(opts, zap.Development())
	}
	// Only add caller in verbose mode
	if cfg.Verbose {
		opts = append(opts, zap.AddCaller())
	}

	return zap.New(core, opts...), nil
}

// Get returns the global logger
func Get() *zap.Logger {
	mu.Lock()
	defer mu.Unlock()

	if globalLogger == nil {
		// Initialize with default config if not initialized
		logger, _ := NewLogger(DefaultConfig())
		globalLogger = logger
	}
	return globalLogger
}

// WithWorkflow returns a logger with workflow context
func WithWorkflow(workflowName, runUUID string) *zap.Logger {
	return Get().With(
		zap.String("workflow", workflowName),
		zap.String("run_uuid", runUUID),
	)
}

// WithStep returns a logger with step context
func WithStep(workflowName, runUUID, stepName string) *zap.Logger {
	return Get().With(
		zap.String("workflow", workflowName),
		zap.String("run_uuid", runUUID),
		zap.String("step", stepName),
	)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// NewFileLogger creates a logger that writes to a specific file
// This is useful for server mode where logs should also go to a workspace file
func NewFileLogger(logFilePath, workflowName, runUUID string) (*zap.Logger, error) {
	if logFilePath == "" {
		return nil, nil
	}

	// Ensure parent directory exists
	if dir := filepath.Dir(logFilePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// File encoder without colors
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     PlainTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	jsonEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
	fileCore := zapcore.NewCore(
		jsonEncoder,
		zapcore.AddSync(file),
		zapcore.DebugLevel, // Capture all levels to file
	)

	return zap.New(fileCore).With(
		zap.String("workflow", workflowName),
		zap.String("run_uuid", runUUID),
	), nil
}

// WithFileOutput creates a combined logger that writes to both the base logger and a file
func WithFileOutput(baseLogger *zap.Logger, logFilePath string) (*zap.Logger, error) {
	if logFilePath == "" || baseLogger == nil {
		return baseLogger, nil
	}

	// Ensure parent directory exists
	if dir := filepath.Dir(logFilePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// File encoder without colors
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     PlainTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncoderConfig),
		zapcore.AddSync(file),
		zapcore.DebugLevel,
	)

	// Combine the base logger core with the file core
	combinedCore := zapcore.NewTee(baseLogger.Core(), fileCore)
	return zap.New(combinedCore), nil
}

// NewStepLogger creates a logger for a specific step that writes to a file
func NewStepLogger(logDir, workflowName, runUUID, stepName string) (*zap.Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	logFile := filepath.Join(logDir, stepName+".log")
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "message",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(file),
		zapcore.DebugLevel,
	)

	return zap.New(core).With(
		zap.String("workflow", workflowName),
		zap.String("run_uuid", runUUID),
		zap.String("step", stepName),
	), nil
}
