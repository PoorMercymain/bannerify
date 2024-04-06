package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PoorMercymain/bannerify/internal/bannerify/config"
	"github.com/PoorMercymain/bannerify/internal/bannerify/handlers"
	"github.com/PoorMercymain/bannerify/internal/bannerify/repository"
	"github.com/PoorMercymain/bannerify/internal/bannerify/service"
	"github.com/PoorMercymain/bannerify/pkg/logger"
	"github.com/caarlos0/env/v6"
	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Logger().Fatalln("Failed to parse env: %v", err)
	}

	fmt.Println("logs/" + cfg.LogFilePath)
	logger.SetLogFile("logs/" + cfg.LogFilePath)

	m, err := migrate.New("file://"+cfg.MigrationsPath, cfg.DSN())
	if err != nil {
		logger.Logger().Fatalln(zap.Error(err))
	}

	err = repository.ApplyMigrations(m)
	if err != nil {
		logger.Logger().Fatalln(zap.Error(err))
	}

	logger.Logger().Infoln("Migrations applied successfully")

	pool, err := repository.GetPgxPool(cfg.DSN())
	if err != nil {
		logger.Logger().Fatalln(zap.Error(err))
	}

	logger.Logger().Infoln("Postgres connection pool created")

	r := repository.NewBanner(repository.NewPostgres(pool))
	s := service.NewBanner(r)
	h := handlers.NewBanner(s)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", h.Ping)

	server := &http.Server{
		Addr:     fmt.Sprintf("%s:%d", cfg.ServiceHost, cfg.ServicePort),
		ErrorLog: log.New(logger.Logger(), "", 0),
		Handler:  mux,
	}

	go func() {
		logger.Logger().Infoln("Server started, listening on port", cfg.ServicePort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger().Fatalln("ListenAndServe failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	logger.Logger().Infoln("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger().Fatalln("Server was forced to shutdown:", zap.Error(err))
	}

	logger.Logger().Infoln("Server was shut down")
}