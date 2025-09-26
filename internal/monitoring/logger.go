package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional context
type Logger struct {
	*zap.Logger
	service string
	version string
}

// LogLevel represents the logging level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// LogFormat represents the logging format
type LogFormat string

const (
	JSONFormat LogFormat = "json"
	TextFormat LogFormat = "text"
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level    LogLevel  `yaml:"level"`
	Format   LogFormat `yaml:"format"`
	Output   string    `yaml:"output"` // stdout, file, both
	FilePath string    `yaml:"file_path"`
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	TraceID     string                 `json:"trace_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Stack       string                 `json:"stack,omitempty"`
	Duration    int64                  `json:"duration_ms,omitempty"`
	HTTPMethod  string                 `json:"http_method,omitempty"`
	HTTPPath    string                 `json:"http_path,omitempty"`
	HTTPStatus  int                    `json:"http_status,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	RemoteAddr  string                 `json:"remote_addr,omitempty"`
}

// NewLogger creates a new structured logger
func NewLogger(service, version string, config LogConfig) (*Logger, error) {
	// Set default values
	if config.Level == "" {
		config.Level = InfoLevel
	}
	if config.Format == "" {
		config.Format = JSONFormat
	}
	if config.Output == "" {
		config.Output = "stdout"
	}

	// Configure zap level
	var zapLevel zapcore.Level
	switch config.Level {
	case DebugLevel:
		zapLevel = zapcore.DebugLevel
	case InfoLevel:
		zapLevel = zapcore.InfoLevel
	case WarnLevel:
		zapLevel = zapcore.WarnLevel
	case ErrorLevel:
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Configure zap encoder
	var encoderConfig zapcore.EncoderConfig
	if config.Format == JSONFormat {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "message"
	encoderConfig.CallerKey = "caller"
	encoderConfig.StacktraceKey = "stack"

	var encoder zapcore.Encoder
	if config.Format == JSONFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure output
	var writeSyncer zapcore.WriteSyncer
	switch config.Output {
	case "stdout":
		writeSyncer = zapcore.AddSync(os.Stdout)
	case "file":
		if config.FilePath == "" {
			config.FilePath = fmt.Sprintf("logs/%s.log", service)
		}
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writeSyncer = zapcore.AddSync(file)
	case "both":
		if config.FilePath == "" {
			config.FilePath = fmt.Sprintf("logs/%s.log", service)
		}
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(file))
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)

	// Create logger with service and version fields
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	logger = logger.With(
		zap.String("service", service),
		zap.String("version", version),
	)

	return &Logger{
		Logger:  logger,
		service: service,
		version: version,
	}, nil
}

// WithContext adds context fields to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract trace ID from context if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		return &Logger{
			Logger:  l.Logger.With(zap.String("trace_id", traceID.(string))),
			service: l.service,
			version: l.version,
		}
	}
	return l
}

// WithUser adds user context to the logger
func (l *Logger) WithUser(userID string) *Logger {
	return &Logger{
		Logger:  l.Logger.With(zap.String("user_id", userID)),
		service: l.service,
		version: l.version,
	}
}

// WithFields adds custom fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return &Logger{
		Logger:  l.Logger.With(zapFields...),
		service: l.service,
		version: l.version,
	}
}

// LogHTTPRequest logs HTTP request details
func (l *Logger) LogHTTPRequest(method, path string, statusCode int, duration time.Duration, userAgent, remoteAddr string) {
	l.Info("HTTP request completed",
		zap.String("http_method", method),
		zap.String("http_path", path),
		zap.Int("http_status", statusCode),
		zap.Duration("duration", duration),
		zap.String("user_agent", userAgent),
		zap.String("remote_addr", remoteAddr),
	)
}

// LogLicenseVerification logs license verification events
func (l *Logger) LogLicenseVerification(licenseKey, product string, valid bool, duration time.Duration) {
	l.Info("License verification",
		zap.String("license_key", licenseKey),
		zap.String("product", product),
		zap.Bool("valid", valid),
		zap.Duration("duration", duration),
	)
}

// LogLicenseActivation logs license activation events
func (l *Logger) LogLicenseActivation(licenseKey, product, machineID string, success bool) {
	l.Info("License activation",
		zap.String("license_key", licenseKey),
		zap.String("product", product),
		zap.String("machine_id", machineID),
		zap.Bool("success", success),
	)
}

// LogUserAction logs user actions
func (l *Logger) LogUserAction(userID, action string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("user_id", userID),
		zap.String("action", action),
	}
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}
	l.Info("User action", fields...)
}

// LogDatabaseOperation logs database operations
func (l *Logger) LogDatabaseOperation(operation, table string, duration time.Duration, success bool) {
	l.Info("Database operation",
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration),
		zap.Bool("success", success),
	)
}

// LogSecurityEvent logs security-related events
func (l *Logger) LogSecurityEvent(eventType, severity string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("severity", severity),
	}
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}
	l.Warn("Security event", fields...)
}

// LogError logs errors with stack trace
func (l *Logger) LogError(err error, message string, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	l.Error(message, allFields...)
}

// LogPanic logs panics and recovers
func (l *Logger) LogPanic(recoverValue interface{}, stack []byte) {
	l.Error("Panic recovered",
		zap.Any("panic_value", recoverValue),
		zap.String("stack", string(stack)),
	)
}

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(operation string, duration time.Duration, memoryBefore, memoryAfter uint64) {
	l.Info("Performance metric",
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Uint64("memory_before_bytes", memoryBefore),
		zap.Uint64("memory_after_bytes", memoryAfter),
		zap.Uint64("memory_delta_bytes", memoryAfter-memoryBefore),
	)
}

// LogAudit logs audit trail events
func (l *Logger) LogAudit(action, resource string, userID string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("audit_action", action),
		zap.String("audit_resource", resource),
		zap.String("user_id", userID),
	}
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}
	l.Info("Audit log", fields...)
}

// LogBusinessEvent logs business-specific events
func (l *Logger) LogBusinessEvent(eventType string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("business_event", eventType),
	}
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}
	l.Info("Business event", fields...)
}

// GetLogEntry creates a structured log entry
func (l *Logger) GetLogEntry(level, message string, fields map[string]interface{}) LogEntry {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Service:   l.service,
		Version:   l.version,
		Message:   message,
		Fields:    fields,
	}

	// Add stack trace for errors
	if level == "error" {
		stack := make([]byte, 1024)
		length := runtime.Stack(stack, false)
		entry.Stack = string(stack[:length])
	}

	return entry
}

// WriteLogEntry writes a structured log entry
func (l *Logger) WriteLogEntry(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		l.Error("Failed to marshal log entry", zap.Error(err))
		return
	}
	fmt.Println(string(data))
}

// Flush flushes any buffered log entries
func (l *Logger) Flush() error {
	return l.Logger.Sync()
}

// Close closes the logger and flushes any buffered entries
func (l *Logger) Close() error {
	return l.Logger.Sync()
}