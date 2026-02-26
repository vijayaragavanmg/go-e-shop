// Package main provides the entry point for the API service.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vijayaragavanmg/learning-go-shop/internal/config"
	"github.com/vijayaragavanmg/learning-go-shop/internal/database"
	"github.com/vijayaragavanmg/learning-go-shop/internal/events"
	"github.com/vijayaragavanmg/learning-go-shop/internal/interfaces"
	"github.com/vijayaragavanmg/learning-go-shop/internal/logger"
	"github.com/vijayaragavanmg/learning-go-shop/internal/providers"
	"github.com/vijayaragavanmg/learning-go-shop/internal/repositories"
	"github.com/vijayaragavanmg/learning-go-shop/internal/server"
	"github.com/vijayaragavanmg/learning-go-shop/internal/services"
)

// @title E-Commerce API
// @version 1.0
// @description A modern e-commerce API built with Go, Gin, and GORM
// @termsOfService http://swagger.io/terms/

// @contact.name   Vijayaragavan
// @contact.url    http://linkedin.com/in/vijayaragavan
// @contact.email  mvijayaragavan@live.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemas http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	log := logger.New()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	mainDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get database connection")
	}

	defer func() {
		if err := mainDB.Close(); err != nil {
			log.Fatal().Err(err).Msg("failed to close mainDB")
		}
	}()

	ctx := context.Background()

	eventPublisher, err := events.NewEventPublisher(ctx, &cfg.AWS)
	if err != nil {
		log.Error().Err(err).Msg("failed to create event publisher")
		return
	}
	gin.SetMode(cfg.Server.GinMode)

	userRepo := repositories.NewUserRepository(db)
	cartRepo := repositories.NewCartRepository(db)
	productRepo := repositories.NewProductRepository(db)
	orderRepo := repositories.NewOrderRepository(db)

	authService := services.NewAuthService(userRepo, cartRepo, cfg, eventPublisher)
	productService := services.NewProductService(productRepo)
	userService := services.NewUserService(userRepo)
	cartService := services.NewCartService(cartRepo, productRepo)
	orderService := services.NewOrderService(orderRepo)

	var uploadProvider interfaces.UploadProvider
	if cfg.Upload.UploadProvider == "s3" {
		uploadProvider = providers.NewS3Provider(cfg, log)
	} else {
		uploadProvider = providers.NewLocalUploadProvider(cfg.Upload.Path, log)
	}
	uploadService := services.NewUploadService(uploadProvider)
	srv := server.New(cfg,
		log,
		authService,
		productService,
		userService, uploadService,
		cartService, orderService)
	router := srv.SetupRoutes()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("starting http server")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to start http server")
		}
	}()
	log.Info().Msg("starting server")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown http server")
		return
	}

}
