package main

import (
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
		panic(err)
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

	slog.Info("starting server", slog.String("addr", cfg.HTTPServer.Address))

	httpServer := http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      wrapped,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout}

	if err := httpServer.ListenAndServe(); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
	}
}
