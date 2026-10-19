package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "indoor-navigation/backend/docs/swagger"
	"indoor-navigation/backend/internal/handler"
	"indoor-navigation/backend/internal/middleware"
)

// Dependencies HTTP 层依赖。
type Dependencies struct {
	Route   *handler.RouteHandler
	Nearby  *handler.NearbyHandler
	Catalog *handler.CatalogHandler
	Health  *handler.HealthHandler
}

// New 构建 Gin Engine 并注册全部接口。
func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	logger := slog.Default()
	engine.Use(middleware.Recovery(logger), middleware.RequestID(), middleware.AccessLog(logger))

	engine.POST("/api/v1/route", deps.Route.Create)
	engine.GET("/api/v1/nearby", deps.Nearby.Get)
	engine.GET("/buildings", deps.Catalog.Buildings)
	engine.GET("/buildings/:code/floors", deps.Catalog.Floors)
	engine.GET("/classrooms", deps.Catalog.Classrooms)
	engine.GET("/health", deps.Health.Get)

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return engine
}
