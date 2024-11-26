package httpkit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/taxfyle/go-httpkit/v5/log"
	"go.uber.org/zap"
)

type ctxkey string

var (
	keyRequestID            ctxkey = "github.com/taxfyle/go-httpkit:request_id"
	DefaultHistogramBuckets        = []float64{100, 250, 500, 1000, 2000, 5000} // milliseconds
)

type Handler interface {
	http.Handler
}

type Server struct {
	mux *http.ServeMux

	cfg ServerConfig

	latencyHistogram *prometheus.HistogramVec
}

type ServerConfig struct {
	MetricsNamespace string
	ServiceName      string
	LatencyBuckets   []float64
	MetricsPath      string
}

func NewServer(mux *http.ServeMux, cfg ServerConfig) (*Server, error) {
	if len(cfg.LatencyBuckets) < 1 {
		cfg.LatencyBuckets = DefaultHistogramBuckets
	}

	if cfg.MetricsNamespace == "" {
		cfg.MetricsNamespace = "httpkit"
	}

	latencyHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: cfg.MetricsNamespace,
		Name:      "http_stats",
		Buckets:   cfg.LatencyBuckets,
		Help:      "Latencies of HTTP requests",
	}, []string{"service", "path", "method", "status"})

	if err := prometheus.DefaultRegisterer.Register(latencyHistogram); err != nil {
		return nil, err
	}

	if cfg.MetricsPath == "" {
		cfg.MetricsPath = "/metrics"
	}

	mux.Handle(fmt.Sprintf("GET %v", cfg.MetricsPath), promhttp.Handler())

	return &Server{
		mux: mux,

		latencyHistogram: latencyHistogram,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	ctx := context.WithValue(r.Context(), keyRequestID, requestID)

	logger := log.FromContext(ctx).With(zap.String("request.id", requestID))

	timeStart := time.Now()

	// This handles the case when the matched route is just a path as well as a method and a path.
	_, muxRoute := s.mux.Handler(r)
	routeSegments := strings.Split(muxRoute, " ")
	path := routeSegments[len(routeSegments)-1]

	lrw := &ResponseWriter{
		ResponseWriter: w,
		status:         200,
	}

	defer func() {
		latency := time.Since(timeStart)

		s.latencyHistogram.WithLabelValues(
			s.cfg.ServiceName,
			path,
			r.Method,
			fmt.Sprintf("%v", lrw.status)).
			Observe(float64(latency.Milliseconds()))

		logger.Sugar().With(
			"http.method", r.Method,
			"http.raw_path", r.URL.Path,
			"http.status", lrw.status,
			"http.response_time", latency.Milliseconds(),
			"http.path", path,
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
