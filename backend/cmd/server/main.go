package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/imageforge/imageforge/internal/config"
	"github.com/imageforge/imageforge/internal/pkg/logger"
	"github.com/imageforge/imageforge/internal/server"
	zapcore "go.uber.org/zap"
)

// @title			ImageForge API
// @version		0.1.0
// @description	Commercial AI image-production workspace.
// @host		localhost:8080
// @BasePath	/v1
func main() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Server.Mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	app, cleanup, err := InitializeApplication(cfg, log)
	if err != nil {
		log.Fatal("failed to initialize application", zapcore.String("error", err.Error()))
	}
	if cleanup != nil {
		defer cleanup()
	}

	// 判断是否启用 TLS（生产环境通常在 Caddy 终止 TLS）
	// Validate TLS configuration: both cert and key must be provided together
	if (cfg.Server.TLSCertFile != "") != (cfg.Server.TLSKeyFile != "") {
		log.Fatal("TLS configuration incomplete: both TLSCertFile and TLSKeyFile must be provided together")
	}
	useTLS := cfg.Server.TLSCertFile != "" && cfg.Server.TLSKeyFile != ""

	addr := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))
	srv := &http.Server{
		Addr:              addr,
		Handler:           app.Router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      time.Duration(cfg.Server.Timeout) * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("ImageForge server starting",
			zapcore.String("addr", addr),
			zapcore.Bool("tls", useTLS),
		)
		var err error
		if useTLS {
			err = srv.ListenAndServeTLS(cfg.Server.TLSCertFile, cfg.Server.TLSKeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server listen failed", zapcore.String("error", err.Error()))
		}
	}()

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zapcore.String("error", err.Error()))
	}
	log.Info("server exited")
}

// app wires the router and cleanup. Populated by Wire via InitializeApplication.
type app struct {
	Router *server.Router
}

// Ensure server package is referenced even before Wire generation completes.
var _ = (*server.Router)(nil)
