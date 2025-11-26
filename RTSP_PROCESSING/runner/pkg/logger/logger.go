package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

// InitLogger инициализирует zap логгер с JSON форматом для Loki
func InitLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	// Направляем обычные логи в stdout, а ошибки в stderr
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Настраиваем формат времени для лучшей совместимости с Loki
	config.EncoderConfig.TimeKey = "ts"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	config.EncoderConfig.MessageKey = "msg"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	// Устанавливаем уровень логирования
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// WithContext возвращает контекст, в который вложен логгер
func WithContext(ctx context.Context, lg *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, lg)
}

// FromContext извлекает логгер из контекста.
// Если его нет — возвращает "noop" логгер.
func FromContext(ctx context.Context) *zap.Logger {
	v := ctx.Value(ctxKey{})
	if v == nil {
		return zap.NewNop()
	}
	if lg, ok := v.(*zap.Logger); ok {
		return lg
	}
	return zap.NewNop()
}
