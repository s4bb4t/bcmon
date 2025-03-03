package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Middleware struct {
	handler http.Handler
	logger  *zap.Logger
}

type wiretapper struct {
	rw http.ResponseWriter

	size   int
	status int
}

func (w *wiretapper) Header() http.Header {
	return w.rw.Header()
}

func (w *wiretapper) Write(b []byte) (n int, err error) {
	n, err = w.rw.Write(b)
	w.size += n
	return
}

func (w *wiretapper) WriteHeader(statusCode int) {
	w.rw.WriteHeader(statusCode)
	w.status = statusCode
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := &wiretapper{
		rw:     w,
		size:   0,
		status: 0,
	}

	start := time.Now()
	m.handler.ServeHTTP(h, r)
	latency := time.Since(start)

	status := h.status
	if status == 0 {
		status = http.StatusOK
	}

	m.logger.Debug("complete handling request",
		zap.Int("status", status),
		zap.String("method", r.Method),
		zap.String("request", r.RequestURI),
		zap.String("remote", r.RemoteAddr),
		zap.Float64("duration", float64(latency.Nanoseconds())/float64(1000)),
		zap.Int("size", h.size),
		zap.String("referer", r.Referer()),
		zap.String("user-agent", r.UserAgent()),
	)
}

func HttpHandler(logger *zap.Logger, h http.Handler) http.Handler {
	return &Middleware{
		handler: h,
		logger:  logger,
	}
}
