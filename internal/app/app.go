// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tarkib.uz/config"
	v1 "tarkib.uz/internal/controller/http/v1"
	"tarkib.uz/internal/usecase"
	"tarkib.uz/internal/usecase/repo"
	"tarkib.uz/pkg/httpserver"
	"tarkib.uz/pkg/logger"
	"tarkib.uz/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	clientOptions := options.Client().ApplyURI("mongodb://news-user:news-password@mongodb-news:27017/news")

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "redis-news:6379", // Redis server address
		Password: "",                // no password set
		DB:       0,                 // use default DB
	})

	pubsubClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Adjust the address as needed
	})

	database := mongoClient.Database("news")

	casbinEnforcer, err := casbin.NewEnforcer(cfg.Casbin.ConfigFilePath, cfg.Casbin.CSVFilePath)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - casbin.NewEnforcer: %w", err))
	}

	// Use case
	authUseCase := usecase.NewAuthUseCase(
		repo.NewAuthRepo(pg),
		cfg,
	)

	adRepo := repo.NewAdRepo(pg)
	adsUseCase := usecase.NewAdUseCase(*adRepo, *cfg)

	categoryUseCase := usecase.NewCategoryUseCase(repo.NewCategoryRepo(pg, database), cfg)

	// HTTP Server
	handler := gin.New()
	v1.NewRouter(handler, l, authUseCase, categoryUseCase, adsUseCase, casbinEnforcer, cfg, client, pubsubClient)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
