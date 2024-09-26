package httpkit

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/taxfyle/go-httpkit/v2/log"
	"go.uber.org/zap"
)

type ctxkey string

var (
	keyRequestID ctxkey = "github.com/taxfyle/go-httpkit:request_id"
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
	requestID := uuid.New().String()
	ctx := context.WithValue(r.Context(), keyRequestID, requestID)

	logger := log.FromContext(ctx).With(zap.String("request.id", requestID))

	timeStart := time.Now()

	lrw := &ResponseWriter{
		ResponseWriter: w,
		status:         200,
	}

	defer func() {
		logger.Sugar().With(
			"http.method", r.Method,
			"http.path", r.URL.Path,
			"http.status", lrw.status,
			"http.response_time", time.Since(timeStart),
		).Info()
	}()

	s.mux.ServeHTTP(lrw, r.WithContext(log.WithContext(ctx, logger)))
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
	requestID, ok := ctx.Value(keyRequestID).(string)
	if !ok {
		requestID = "UNSET"
	}
	logger := log.FromContext(ctx).Sugar()

	resp := struct {
		Error     string `json:"error"`
		RequestID string `json:"request_id"`
	}{
		Error:     err.Error(),
		RequestID: requestID,
	}

	buf, err := json.Marshal(resp)
	if err != nil {
		logger.With("error", err).Error("unable to marshal error response json")
		status = http.StatusInternalServerError
	}

	rw.WriteHeader(status)
	rw.Write(buf)
}
