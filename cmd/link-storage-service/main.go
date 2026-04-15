package main

import (
	"link-storage-service/internal/config"
	json_middleware "link-storage-service/internal/handler/json-middleware"
	"link-storage-service/internal/http-server/handlers/link/get"
	"link-storage-service/internal/http-server/handlers/link/save"
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

	saveHandler := save.New(storage)
	getHandler := get.New(storage)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /links", saveHandler)
	mux.HandleFunc("GET /links/{short_code}", getHandler)
	wrapped := json_middleware.JsonMiddleware(mux)

	slog.Info("starting server", slog.String("addr", cfg.HTTPServer.Address))

	if err := http.ListenAndServe(cfg.HTTPServer.Address, wrapped); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
	}
}
