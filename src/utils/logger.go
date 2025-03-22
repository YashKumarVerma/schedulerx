package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

var once sync.Once

// StandardLogger enforces specific log message formats.
type StandardLogger struct {
	*zap.SugaredLogger
}

// IntegerLevelEncoder returns custom encoder for level field.
func IntegerLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt8((int8(l) + 3) * 10)
}

var appLogger *StandardLogger

// NewLogger creates a new application logger.
func NewLogger() *StandardLogger {
	var cfg zap.Config
	outputLevel := zap.InfoLevel
	levelEnv := os.Getenv("LOG_LEVEL")
	if levelEnv != "" {
		levelFromEnv, err := zapcore.ParseLevel(levelEnv)
		if err != nil {
			log.Println(
				fmt.Errorf("invalid level, defaulting to INFO: %w", err),
			)
		}
		outputLevel = levelFromEnv
	}
	var DgnEnv = os.Getenv("DGN")
	if DgnEnv != "local" {
		cfg = zap.NewProductionConfig()
		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stdout"}
		cfg.InitialFields = map[string]any{"name": "hacky scheduler"}
		cfg.EncoderConfig.EncodeLevel = IntegerLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.TimeKey = "time"
		cfg.Level = zap.NewAtomicLevelAt(outputLevel)
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return &StandardLogger{SugaredLogger: logger.Sugar()}
}

func GetAppLogger(ctx context.Context) *StandardLogger {
	once.Do(func() {
		appLogger = NewLogger()
	})
	return LoggerFromCtx(ctx)
}

func GetChildLogger(parent *StandardLogger, childContext map[string]string) *StandardLogger {
	zapFields := make([]any, 0)
	for k, v := range childContext {
		zapFields = append(zapFields, zap.String(k, v))
	}
	return &StandardLogger{parent.With(zapFields...)}
}

// LoggerFromCtx returns the Logger associated with the ctx. If no logger
// is associated, the default logger is returned, unless it is nil
// in which case a disabled logger is returned.
func LoggerFromCtx(ctx context.Context) *StandardLogger {
	if l, ok := ctx.Value(ctxKey{}).(*StandardLogger); ok {
		return l
	} else if l := appLogger; l != nil {
		return l
	}
	return &StandardLogger{zap.NewNop().Sugar()}
}

// LoggerWithCtx returns a copy of ctx with the Logger attached.
func LoggerWithCtx(ctx context.Context, l *StandardLogger) context.Context {
	if lp, ok := ctx.Value(ctxKey{}).(*StandardLogger); ok {
		if lp == l {
			// Do not store same logger.
			return ctx
		}
	}

	return context.WithValue(ctx, ctxKey{}, l)
}
