package httpkit

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/taxfyle/go-httpkit/log"
)

type Handler interface {
	http.Handler
}

type Server struct {
	mux *http.ServeMux
}

func NewServer(mux *http.ServeMux) *Server {
	return &Server{
		mux: mux,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, logger := log.NewContext(r.Context(), nil)

	timeStart := time.Now()

	lrw := &ResponseWriter{
		ResponseWriter: w,
		status:         200,
	}

	defer func() {
		logger.WithFields(
			"http.method", r.Method,
			"http.path", r.URL.Path,
			"http.status", lrw.status,
			"http.response_time", time.Since(timeStart),
		).Info()
	}()

	s.mux.ServeHTTP(lrw, r.WithContext(ctx))
}

type ResponseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func DefaultErrorHandler(ctx context.Context, rw http.ResponseWriter, err error, status int) {
	logger := log.FromContext(ctx)

	resp := struct {
		Error     string `json:"error"`
		RequestID string `json:"request_id"`
	}{
		Error:     err.Error(),
		RequestID: logger.ID,
	}

	buf, err := json.Marshal(resp)
	if err != nil {
		logger.With("error", err).Error("unable to marshal error response json")
		status = http.StatusInternalServerError
	}

	rw.WriteHeader(status)
	rw.Write(buf)
}
