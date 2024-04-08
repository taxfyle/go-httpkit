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

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	log.BaseLogger = logger.Sugar()

	ctx, cancel := context.WithCancel(context.Background())

	log.BaseLogger.Info("booting server")

	s := http.Server{
		Addr: ":9999",
		Handler: &httpkit.Server{
			Handlers: []httpkit.Handler{
				&health.Handler{
					Path: "/health",
				},
			},
		},
	}

	go func() {
		logger.Info("running server")

		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.BaseLogger.With("error", err).Fatal("error serving http")
		}
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM)

		<-c
		logger.Info("received shutdown signal, cleaning up")
		if err := s.Shutdown(ctx); err != nil {
			log.BaseLogger.With("error", err).Error("unable to cleanly exit server")
		}
		cancel()
	}()

	<-ctx.Done()
}
