package main

import (
	"log/slog"
	"os"

	"absensi-cn-api/internal/app"
)

func main() {
	server, err := app.New()
	if err != nil {
		slog.Error("failed to bootstrap api", "error", err)
		os.Exit(1)
	}

	if err := server.Run(); err != nil {
		slog.Error("failed to run api", "error", err)
		os.Exit(1)
	}
}
