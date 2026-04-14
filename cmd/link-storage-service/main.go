package main

import (
	"link-storage-service/internal/config"
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

	mux := http.NewServeMux()
	mux.HandleFunc("/links", saveHandler)

	slog.Info("starting server", slog.String("addr", cfg.HTTPServer.Address))

	if err := http.ListenAndServe(cfg.HTTPServer.Address, mux); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
	}
}
