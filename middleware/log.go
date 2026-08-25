package middleware

import (
	"time"

	"github.com/allentom/haruka"
	"github.com/sirupsen/logrus"
)

// RequestLoggerMiddleware logs one line per request including status
// code, duration, client ip and request id. It implements both
// haruka.Middleware and haruka.ResponseMiddleware.
type RequestLoggerMiddleware struct {
	Logger *logrus.Logger
}

func (r *RequestLoggerMiddleware) logger() *logrus.Logger {
	if r.Logger != nil {
		return r.Logger
	}
	r.Logger = logrus.New()
	return r.Logger
}

func (r *RequestLoggerMiddleware) OnRequest(ctx *haruka.Context) {
}

func (r *RequestLoggerMiddleware) OnResponse(ctx *haruka.Context) {
	requestID, _ := ctx.Param["requestId"].(string)
	fields := logrus.Fields{
		"method":   ctx.Request.Method,
		"path":     ctx.Request.URL.Path,
		"status":   ctx.StatusCode(),
		"duration": time.Since(ctx.StartTime()).String(),
		"ip":       ctx.ClientIP(),
	}
	if requestID != "" {
		fields["requestId"] = requestID
	}
	r.logger().WithFields(fields).Info()
}

func NewLoggerMiddleware() *RequestLoggerMiddleware {
	return &RequestLoggerMiddleware{
		Logger: logrus.New(),
	}
}

var _ haruka.ResponseMiddleware = (*RequestLoggerMiddleware)(nil)
