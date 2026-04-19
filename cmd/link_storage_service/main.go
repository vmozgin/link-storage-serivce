package main

import (
	"link-storage-service/internal/cache"
	"link-storage-service/internal/config"
	modelLink "link-storage-service/internal/domain/link"
	"link-storage-service/internal/http/handlers/link"
	"link-storage-service/internal/http/middleware/json"
	"link-storage-service/internal/http/middleware/logging"
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
	linkCache := cache.NewCache[modelLink.SimpleLink]()

	saveHandler := link.Create(storage)
	getHandler := link.Get(storage, linkCache)
	deleteHandler := link.Delete(storage, linkCache)
	statsHandler := link.Stats(storage)
	getAllHandler := link.GetAll(storage)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /links", saveHandler)
	mux.HandleFunc("GET /links/{short_code}", getHandler)
	mux.HandleFunc("DELETE /links/{short_code}", deleteHandler)
	mux.HandleFunc("GET /links/{short_code}/stats", statsHandler)
	mux.HandleFunc("GET /links", getAllHandler)
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
