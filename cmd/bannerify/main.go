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

	"github.com/caarlos0/env/v6"
	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"

	"github.com/PoorMercymain/bannerify/internal/bannerify/config"
	"github.com/PoorMercymain/bannerify/internal/bannerify/handlers"
	"github.com/PoorMercymain/bannerify/internal/bannerify/middleware"
	"github.com/PoorMercymain/bannerify/internal/bannerify/repository"
	"github.com/PoorMercymain/bannerify/internal/bannerify/service"
	"github.com/PoorMercymain/bannerify/pkg/logger"
)

func main() {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Logger().Fatalln("Failed to parse env: %v", err)
	}

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

	pg := repository.NewPostgres(pool)

	r := repository.NewBanner(pg)
	s := service.NewBanner(r)
	h := handlers.NewBanner(s)

	ar := repository.NewAuthorization(pg)
	as := service.NewAuthorization(ar)
	ah := handlers.NewAuthorization(as, cfg.JWTKey)

	mux := http.NewServeMux()

	mux.Handle("GET /ping", middleware.Log(middleware.AdminRequired(http.HandlerFunc(h.Ping), ah.JWTKey)))

	mux.Handle("POST /register", middleware.Log(http.HandlerFunc(ah.Register)))
	mux.Handle("POST /aquire-token", middleware.Log(http.HandlerFunc(ah.LogIn)))

	mux.Handle("GET /user_banner", middleware.Log(middleware.AuthorizationRequired(http.HandlerFunc(h.GetBanner), ah.JWTKey)))
	mux.Handle("GET /banner", middleware.Log(middleware.AdminRequired(http.HandlerFunc(h.ListBanners), ah.JWTKey)))
	mux.Handle("GET /banner_versions/{banner_id}", middleware.Log(middleware.AdminRequired(http.HandlerFunc(h.ListVersions), ah.JWTKey)))
	mux.Handle("PATCH /banner_versions/choose/{banner_id}", middleware.Log(middleware.AdminRequired(http.HandlerFunc(h.ChooseVersion), ah.JWTKey)))
	mux.Handle("POST /banner", middleware.Log(middleware.AdminRequired(http.HandlerFunc(h.CreateBanner), ah.JWTKey)))

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
