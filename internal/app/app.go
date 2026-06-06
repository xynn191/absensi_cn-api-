package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"absensi-cn-api/internal/config"
	"absensi-cn-api/internal/database"
	"absensi-cn-api/internal/router"
)

type App struct {
	config *config.Config
	server *http.Server
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	db, err := database.NewMySQL(cfg)
	if err != nil {
		return nil, err
	}

	engine := router.New(cfg, db)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{
		config: cfg,
		server: server,
	}, nil
}

func (a *App) Run() error {
	slog.Info(
		"starting absensi cn api",
		"port", a.config.Port,
		"env", a.config.Environment,
		"db_enabled", a.config.Database.Enabled,
	)

	return a.server.ListenAndServe()
}
