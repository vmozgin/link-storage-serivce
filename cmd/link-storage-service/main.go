package main

import (
	cache2 "link-storage-service/internal/cache"
	"link-storage-service/internal/config"
	"link-storage-service/internal/handler/middleware/json"
	"link-storage-service/internal/handler/middleware/logging"
	delete2 "link-storage-service/internal/http-server/handlers/link/delete"
	"link-storage-service/internal/http-server/handlers/link/get"
	get_all "link-storage-service/internal/http-server/handlers/link/get-all"
	"link-storage-service/internal/http-server/handlers/link/save"
	"link-storage-service/internal/http-server/handlers/link/stats"
	"link-storage-service/internal/model/link"
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
	linkCache := cache2.NewCache[link.SimpleLink]()

	saveHandler := save.New(storage)
	getHandler := get.New(storage, linkCache)
	deleteHandler := delete2.New(storage, linkCache)
	statsHandler := stats.New(storage)
	getAllHandler := get_all.New(storage)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /links", saveHandler)
	mux.HandleFunc("GET /links/{short_code}", getHandler)
	mux.HandleFunc("DELETE /links/{short_code}", deleteHandler)
	mux.HandleFunc("GET /links/{short_code}/stats", statsHandler)
	mux.HandleFunc("GET /links", getAllHandler)
	wrapped := json.JsonMiddleware(mux)
	wrapped = logging.LoggingMiddleware(wrapped)

	slog.Info("starting server", slog.String("addr", cfg.HTTPServer.Address))

	if err := http.ListenAndServe(cfg.HTTPServer.Address, wrapped); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
	}
}
