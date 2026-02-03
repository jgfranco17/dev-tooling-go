package logging

import (
	"context"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

type contextLogKey string

const contextKey contextLogKey = "logger"

func New(stream io.Writer, level logrus.Level) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(stream)
	logger.SetLevel(level)

	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:          false,
		PadLevelText:           true,
		QuoteEmptyFields:       true,
		FullTimestamp:          true,
		DisableSorting:         true,
		DisableLevelTruncation: true,
		TimestampFormat:        time.DateTime,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "@level",
			logrus.FieldKeyMsg:   "@message",
		},
	})
	return logger
}

func AddToContext(ctx context.Context, logger *logrus.Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}

func FromContext(ctx context.Context) *logrus.Logger {
	if logger, ok := ctx.Value(contextKey).(*logrus.Logger); ok {
		return logger
	}
	panic("no logger set in context")
}
