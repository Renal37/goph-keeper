package middleware

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
)

// InterceptorLogger возвращает реализацию `logging.Logger` для использования
// с gRPC логирующими перехватчиками. Он отображает данный контекст, уровень
// логирования и сообщение на соответствующую функцию логирования в предоставленном
// `zap.Logger`, преобразуя поля из API `logging.Logger` в экземпляры `zap.Field`.
func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		//nolint:gomnd // Это допустимое число
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			logger.With(zap.Any("lvl", lvl)).Error("неизвестный уровень")
			logger.With(zap.String("msg", msg)).Warn("не удалось обработать сообщение")
		}
	})
}
