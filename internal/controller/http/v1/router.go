// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Swagger docs
	"tarkib.uz/config"
	_ "tarkib.uz/docs"
	"tarkib.uz/internal/controller/middleware"
	"tarkib.uz/internal/usecase"
	"tarkib.uz/pkg/logger"
	tokens "tarkib.uz/pkg/token"
)

// NewRouter -.
// Swagger spec:
// @title       news back-end
// @description Backend - Nodirbek No'monov     TG: https://t.me/alwaysgolang
// @version     1.0
// @BasePath    /v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func NewRouter(handler *gin.Engine, l logger.Interface, authUseCase *usecase.AuthUseCase, categoryUseCase *usecase.CategoryUseCase, adUseCase *usecase.AdUseCase, enforcer *casbin.Enforcer, cfg *config.Config, redisClient *redis.Client, pubSubClient *redis.Client) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	jwtHandler := tokens.JWTHandler{
		SigninKey: cfg.Casbin.SigningKey,
	}
	handler.Use(middleware.NewAuthorizer(enforcer, jwtHandler, cfg, l))

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Authorization", "Content-Type", "Accept"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowBrowserExtensions = true

	handler.Use(cors.New(corsConfig))

	// Swagger
	url := ginSwagger.URL("swagger/doc.json")
	handler.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// K8s probe
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/v1")
	{
		newAuthRoutes(h, authUseCase, l)
		newCategoryRoutes(h, *categoryUseCase, l)
		newAdRoutes(h, *adUseCase, l)
	}
}
