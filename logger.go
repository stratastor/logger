package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

// Logger is an alias for *slog.Logger to simplify usage and allow method chaining.
type Logger = *slog.Logger

// Config holds the configuration for the logger.
type Config struct {
	LogLevel     string
	SentryDSN    string
	EnableSentry bool
}

// New initializes a new Logger based on the provided configuration.
func New(config Config) (Logger, error) {
	var level slog.Level
	switch config.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	sentryHandler := &sentryHandler{
		next:        nil,
		minLogLevel: slog.LevelWarn,
	}

	combinedHandler := &combinedHandler{
		jsonHandler:   jsonHandler,
		sentryHandler: sentryHandler,
	}

	var logger *slog.Logger
	if config.EnableSentry && config.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              config.SentryDSN,
			EnableTracing:    true,
			TracesSampleRate: 0.05,
		}); err != nil {
			return nil, fmt.Errorf("sentry.Init failed: %s", err)
		}
		defer sentry.Flush(2 * time.Second)
		logger = slog.New(combinedHandler)
	} else {
		logger = slog.New(jsonHandler)
	}

	return logger, nil
}

// NewTag initializes a new Logger with a specific tag added to its context.
func NewTag(config Config, tag string) (Logger, error) {
	logger, err := New(config)
	if err != nil {
		return nil, err
	}

	// Add the tag to the logger's context or configuration if needed
	logger = logger.With(slog.String("tag", tag))

	return logger, nil
}

// sentryHandler is a custom slog.Handler that sends log records to Sentry.
type sentryHandler struct {
	next        slog.Handler
	minLogLevel slog.Level
}

// Handle processes the log record and sends it to Sentry if the log level is high enough.
func (h *sentryHandler) Handle(ctx context.Context, record slog.Record) error {
	// Check if the record's log level meets the minimum level to send to Sentry
	if record.Level < h.minLogLevel {
		if h.next != nil {
			return h.next.Handle(ctx, record) // Pass to the next handler without sending to Sentry
		}
	}

	// Prepare attributes as context for Sentry
	attrs := map[string]interface{}{}
	record.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	// Capture log message as a Sentry event
	sentry.WithScope(func(scope *sentry.Scope) {
		for k, v := range attrs {
			scope.SetExtra(k, v)
		}
		scope.SetLevel(slogToSentryLevel(record.Level)) // Map slog level to Sentry level
		sentry.CaptureMessage(record.Message)
	})

	if h.next != nil {
		return h.next.Handle(ctx, record) // Optionally pass to the next handler
	}
	return nil
}

// Helper function to map slog levels to Sentry levels
func slogToSentryLevel(level slog.Level) sentry.Level {
	switch level {
	case slog.LevelDebug:
		return sentry.LevelDebug
	case slog.LevelInfo:
		return sentry.LevelInfo
	case slog.LevelWarn:
		return sentry.LevelWarning
	case slog.LevelError:
		return sentry.LevelError
	default:
		return sentry.LevelInfo
	}
}

// Enabled determines if the handler is enabled for the given log level.
func (h *sentryHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLogLevel
}

// WithAttrs returns a new handler with the given attributes.
func (h *sentryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &sentryHandler{
		next:        h.next,
		minLogLevel: h.minLogLevel,
	}
}

// WithGroup returns a new handler with the given group name.
func (h *sentryHandler) WithGroup(name string) slog.Handler {
	return &sentryHandler{
		next:        h.next,
		minLogLevel: h.minLogLevel,
	}
}

// combinedHandler is a custom slog.Handler that combines JSON and Sentry handlers.
type combinedHandler struct {
	jsonHandler   slog.Handler
	sentryHandler slog.Handler
}

// Handle processes the log record using both JSON and Sentry handlers.
func (h *combinedHandler) Handle(ctx context.Context, record slog.Record) error {
	// First, handle the log with the JSON handler
	err := h.jsonHandler.Handle(ctx, record)
	if err != nil {
		return err
	}

	// Then, handle the log with the Sentry handler
	return h.sentryHandler.Handle(ctx, record)
}

// Enabled determines if the handler is enabled for the given log level.
func (h *combinedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.jsonHandler.Enabled(ctx, level) || h.sentryHandler.Enabled(ctx, level)
}

// WithAttrs returns a new combined handler with the given attributes.
func (h *combinedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &combinedHandler{
		jsonHandler:   h.jsonHandler.WithAttrs(attrs),
		sentryHandler: h.sentryHandler.WithAttrs(attrs),
	}
}

// WithGroup returns a new combined handler with the given group name.
func (h *combinedHandler) WithGroup(name string) slog.Handler {
	return &combinedHandler{
		jsonHandler:   h.jsonHandler.WithGroup(name),
		sentryHandler: h.sentryHandler.WithGroup(name),
	}
}
