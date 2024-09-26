package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/taxfyle/go-httpkit"
	"github.com/taxfyle/go-httpkit/health"
	"github.com/taxfyle/go-httpkit/log"
	"go.uber.org/zap"
)

type echoHandler struct {
}

func (h *echoHandler) Get(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context()).Sugar()
	logger.Debug("handling echo request")

	route := r.URL.Path
	logger.Debugf("handling route %s", route)

	id := r.PathValue("id")
	logger.Infof("GET id %v", id)

	w.WriteHeader(http.StatusOK)
}

func (h *echoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context()).Sugar()
	logger.Debug("handling DELETE echo request")

	route := r.URL.Path
	logger.Debugf("handling route %s", route)

	id := r.PathValue("id")
	logger.Infof("DELETE id %v", id)
}

func main() {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer l.Sync()
	log.BaseLogger = l

	ctx, cancel := context.WithCancel(context.Background())
	logger := log.FromContext(ctx).Sugar().With("thread", "main")

	logger.Info("booting server")

	echoHandler := &echoHandler{}
	healthHandler := &health.Handler{}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/readiness", healthHandler.GetReadiness)
	mux.HandleFunc("GET /echo/{id}", echoHandler.Get)
	mux.HandleFunc("DELETE /echo/{id}", echoHandler.Delete)

	s := http.Server{
		Addr:    ":9999",
		Handler: httpkit.NewServer(mux),
	}

	go func() {
		logger.Info("running server")

		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			logger.With("error", err).Fatal("error serving http")
		}
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM)

		<-c
		logger.Info("received shutdown signal, cleaning up")
		if err := s.Shutdown(ctx); err != nil {
			logger.With("error", err).Error("unable to cleanly exit server")
		}
		cancel()
	}()

	<-ctx.Done()
}
