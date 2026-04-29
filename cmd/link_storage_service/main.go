package main

import (
	"context"
	"errors"
	"link-storage-service/internal/cache"
	"link-storage-service/internal/config"
	"link-storage-service/internal/http/handlers/link"
	"link-storage-service/internal/http/middleware/json"
	"link-storage-service/internal/http/middleware/logging"
	"link-storage-service/internal/service"
	"link-storage-service/internal/storage/postgres"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()

	storage, err := postgres.New(cfg.Storage)
	if err != nil {
		slog.Error("failed to init storage", "err", err)
		os.Exit(1)
	}
	redisCache, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	linkService := service.NewLinkService(storage, redisCache)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /links", link.Create(linkService))
	mux.HandleFunc("GET /links/{short_code}", link.Get(linkService))
	mux.HandleFunc("DELETE /links/{short_code}", link.Delete(linkService))
	mux.HandleFunc("GET /links/{short_code}/stats", link.Stats(linkService))
	mux.HandleFunc("GET /links", link.GetAll(linkService))
	wrapped := json.JsonMiddleware(mux)
	wrapped = logging.LoggingMiddleware(wrapped)

	httpServer := http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      wrapped,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err,
			http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("server started", "addr", cfg.HTTPServer.Address)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	}

	if err := storage.Close(); err != nil {
		slog.Error("storage close failed", "error", err)
	}
	if err := redisCache.Close(); err != nil {
		slog.Error("redis close failed", "error", err)
	}

	slog.Info("server stopped")

}
