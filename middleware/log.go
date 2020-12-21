package middleware

import (
	"github.com/allentom/haruka"
	"github.com/sirupsen/logrus"
)

type RequestLoggerMiddleware struct {
	Logger *logrus.Logger
}

func (r *RequestLoggerMiddleware) OnRequest(ctx *haruka.Context) {
	r.Logger.WithFields(logrus.Fields{
		"method": ctx.Request.Method,
		"path":   ctx.Request.URL.Path,
		"host":   ctx.Request.Host,
	}).Info()
}
func NewLoggerMiddleware() *RequestLoggerMiddleware {
	return &RequestLoggerMiddleware{
		Logger: logrus.New(),
	}
}
