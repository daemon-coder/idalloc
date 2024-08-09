package log_infra

import (
	"strings"
	"sync"
	"time"

	"github.com/daemon-coder/idalloc/definition"
	e "github.com/daemon-coder/idalloc/definition/errors"
	threadLocal "github.com/daemon-coder/idalloc/infrastructure/threadlocal_infra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var loggerLoadOnce sync.Once

func GetLogger() *zap.SugaredLogger {
	if Logger == nil {
		InitZapLogger()
	}
	return Logger.Sugar()
}

func InitZapLogger() *zap.Logger {
	if Logger != nil {
		return Logger
	}
	loggerLoadOnce.Do(func() {
		config := zap.Config{
			Encoding:         "console",
			Development:      false,
			Level:            zap.NewAtomicLevelAt(transformLogLevel(definition.Cfg.LogLevel)),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				ConsoleSeparator: "|",
				MessageKey:       "msg",
				LevelKey:         "level",
				TimeKey:          "timestamp",
				CallerKey:        "caller",
				StacktraceKey:    "stacktrace",
				LineEnding:       zapcore.DefaultLineEnding,
				EncodeLevel:      zapcore.CapitalLevelEncoder,
				EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
					enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
				},
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		}

		config.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(caller.TrimmedPath())
			// Customize the encoder to include request_id in the log prefix
			enc.AppendString(threadLocal.GetTraceId())
		}

		var err error
		Logger, err = config.Build()
		if err != nil {
			e.FromStdError(err).Panic()
		}
	})
	return Logger
}

func transformLogLevel(level string) zapcore.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "DPANIC":
		return zapcore.DPanicLevel
	case "PANIC":
		return zapcore.PanicLevel
	case "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func LogError(err error, msg string, kvArgs ...interface{}) {
	baseError := e.FromStdError(err)
	switch {
	case baseError.Type == e.OK:
		return
	case baseError.Type >= e.CriticalErrorType:
		GetLogger().Errorw(msg, kvArgs...)
	case baseError.Type >= e.RateLimitErrorType:
		GetLogger().Warnw(msg, kvArgs...)
	default:
		GetLogger().Infow(msg, kvArgs...)
	}
}
