package logger

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	LevelDebug string = "debug"
	LevelInfo  string = "info"
	LevelWarn  string = "warn"
	LevelError string = "error"
	LevelFatal string = "fatal"
	LevelPanic string = "panic"
)

type Config struct {
	IsFile       bool
	FilePath     string
	Level        string
	CustomFields map[string]interface{}
}

type Dependency struct {
	log      *logrus.Logger
	logEntry *logrus.Entry
}

type Logger interface {
	Debug(ctx context.Context, v ...interface{})
	Info(ctx context.Context, v ...interface{})
	Warn(ctx context.Context, v ...interface{})
	Error(ctx context.Context, v ...interface{})
	Fatal(ctx context.Context, v ...interface{})
	Panic(ctx context.Context, v ...interface{})
}

func New(c *Config) Logger {
	log := logrus.New()
	if c.IsFile {
		f, err := os.OpenFile(c.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o755)
		if err != nil {
			panic(err)
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stdout)
	}
	logEntry := log.WithFields(c.CustomFields)
	log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	setLevel(log, c.Level)
	return &Dependency{
		log:      log,
		logEntry: logEntry,
	}
}

func setLevel(log *logrus.Logger, level string) {
	switch level {
	case LevelDebug:
		log.SetLevel(logrus.DebugLevel)
	case LevelInfo:
		log.SetLevel(logrus.InfoLevel)
	case LevelWarn:
		log.SetLevel(logrus.WarnLevel)
	case LevelError:
		log.SetLevel(logrus.ErrorLevel)
	case LevelFatal:
		log.SetLevel(logrus.FatalLevel)
	case LevelPanic:
		log.SetLevel(logrus.PanicLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

func (d *Dependency) buildContextField(ctx context.Context) *logrus.Entry {
	req := ctx.Value(0)
	if req != nil {
		reqTransform := req.(*http.Request)
		d.logEntry = d.logEntry.WithFields(map[string]interface{}{
			"request_id":  reqTransform.Header.Get("x-request-id"),
			"method":      reqTransform.Method,
			"scheme":      reqTransform.Header.Get("x-request-scheme"),
			"client_ip":   reqTransform.Header.Get("x-forwarded-for"),
			"path":        reqTransform.URL.Path,
			"user_agent":  reqTransform.UserAgent(),
			"remote_addr": reqTransform.RemoteAddr,
		})
	}
	return d.logEntry
}

func (l *Dependency) Debug(ctx context.Context, v ...interface{}) {
	l.buildContextField(ctx).Debug(v...)
}

func (l *Dependency) Info(ctx context.Context, v ...interface{}) {
	l.buildContextField(ctx).Info(v...)
}

func (l *Dependency) Warn(ctx context.Context, v ...interface{}) {
	l.buildContextField(ctx).Warn(v...)
}

func (l *Dependency) Error(ctx context.Context, v ...interface{}) {
	l.buildContextField(ctx).Error(v...)
}

func (l *Dependency) Fatal(ctx context.Context, v ...interface{}) {
	l.buildContextField(ctx).Fatal(v...)
}

func (l *Dependency) Panic(ctx context.Context, v ...interface{}) {
	l.buildContextField(ctx).Panic(v...)
}
