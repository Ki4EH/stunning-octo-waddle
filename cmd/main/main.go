package main

import (
	"context"
	"errors"
	"github.com/Ki4EH/stunning-octo-waddle/internal/api"
	"github.com/Ki4EH/stunning-octo-waddle/internal/config"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db"
	"github.com/Ki4EH/stunning-octo-waddle/internal/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log, err := logger.NewZapLogger(cfg.Environment)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	database, err := db.NewPostgresDB(cfg)
	if err != nil {
		panic("failed to create database connection: " + err.Error())
	}

	e := echo.New()

	// Инициализируем пути для API
	api.InitRoutes(e, database, &log)

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err = e.Start(":" + cfg.ServerPort); err != nil && !errors.Is(err, http.ErrServerClosed) {
			defer database.Close()
			log.Error("shutting down the server", zap.Error(err))
			syscall.Exit(1)
		}
	}()

	<-graceCh

	// Graceful shutdown сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = e.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}

	log.Info("server exiting")
}
