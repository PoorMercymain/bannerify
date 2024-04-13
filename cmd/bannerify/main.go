package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	redis, err := repository.NewCache(cfg.CachePort)
	if err != nil {
		logger.Logger().Fatalln(zap.Error(err))
	}

	var wg sync.WaitGroup

	pingProviderRepository := repository.NewPingProvider(pg)
	getterRepository := repository.NewGetter(pg, redis)
	versionerRepository := repository.NewVersioner(pg)
	creatorRepository := repository.NewCreator(pg)
	updaterRepository := repository.NewUpdater(pg)
	deleterRepository := repository.NewDeleter(pg, &wg, cfg.DeleteWorkersAmount)

	pingProviderService := service.NewPingProvider(pingProviderRepository)
	getterService := service.NewGetter(getterRepository)
	versionerService := service.NewVersioner(versionerRepository)
	creatorService := service.NewCreator(creatorRepository)
	updaterService := service.NewUpdater(updaterRepository)
	deleterService := service.NewDeleter(deleterRepository)

	pingProviderHandler := handlers.NewPingProvider(pingProviderService)
	getterHandler := handlers.NewGetter(getterService)
	versionerHandler := handlers.NewVersioner(versionerService)
	creatorHandler := handlers.NewCreator(creatorService)
	updaterHandler := handlers.NewUpdater(updaterService)
	deleterHandler := handlers.NewDeleter(deleterService)

	authRepository := repository.NewAuthorization(pg)
	authService := service.NewAuthorization(authRepository)
	authHandler := handlers.NewAuthorization(authService, cfg.JWTKey)

	deleteCtx, cancelDeleteCtx := context.WithCancel(context.Background())

	mux := http.NewServeMux()

	mux.Handle("GET /ping", middleware.Log(middleware.AdminRequired(http.HandlerFunc(pingProviderHandler.Ping), authHandler.JWTKey)))

	mux.Handle("POST /register", middleware.Log(http.HandlerFunc(authHandler.Register)))
	mux.Handle("POST /acquire-token", middleware.Log(http.HandlerFunc(authHandler.LogIn)))

	mux.Handle("GET /user_banner", middleware.Log(middleware.ProvideIsAdmin(getterHandler.GetBanner, authHandler.JWTKey)))
	mux.Handle("GET /banner", middleware.Log(middleware.AdminRequired(http.HandlerFunc(getterHandler.ListBanners), authHandler.JWTKey)))
	mux.Handle("GET /banner_versions/{banner_id}", middleware.Log(middleware.AdminRequired(http.HandlerFunc(versionerHandler.ListVersions), authHandler.JWTKey)))
	mux.Handle("PATCH /banner_versions/choose/{banner_id}", middleware.Log(middleware.AdminRequired(http.HandlerFunc(versionerHandler.ChooseVersion), authHandler.JWTKey)))
	mux.Handle("POST /banner", middleware.Log(middleware.AdminRequired(http.HandlerFunc(creatorHandler.CreateBanner), authHandler.JWTKey)))
	mux.Handle("PATCH /banner/{id}", middleware.Log(middleware.AdminRequired(http.HandlerFunc(updaterHandler.UpdateBanner), authHandler.JWTKey)))
	mux.Handle("DELETE /banner/{id}", middleware.Log(middleware.AdminRequired(http.HandlerFunc(deleterHandler.DeleteBannerByID), authHandler.JWTKey)))
	mux.Handle("DELETE /banner", middleware.Log(middleware.AdminRequired(http.HandlerFunc(deleterHandler.DeleteBannerByTagOrFeature(deleteCtx, &wg)), authHandler.JWTKey)))

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

	waitGroupChan := make(chan struct{})
	go func() {
		wg.Wait()
		waitGroupChan <- struct{}{}
	}()

	select {
	case <-waitGroupChan:
		logger.Logger().Infoln("All delete goroutines successfully finished")
	case <-time.After(time.Second * 3):
		cancelDeleteCtx()
		logger.Logger().Infoln("Some of delete goroutines have not completed their job due to shutdown timeout")
	}

	logger.Logger().Infoln("Server was shut down")
}
