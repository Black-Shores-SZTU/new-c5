package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"indoor-navigation/backend/internal/cache"
	"indoor-navigation/backend/internal/database"
	"indoor-navigation/backend/internal/graph"
	"indoor-navigation/backend/internal/handler"
	"indoor-navigation/backend/internal/logging"
	"indoor-navigation/backend/internal/repository"
	"indoor-navigation/backend/internal/route"
	"indoor-navigation/backend/internal/router"
	"indoor-navigation/backend/internal/service"
)

// @title C5 Indoor Navigation API
// @version 1.0
// @description C5 教学楼室内导航 MVP 接口
// @BasePath /
// @schemes http

// main 负责启动 HTTP 服务，并在启动阶段把楼栋数据加载为只读内存图。
// 校验或构建失败时直接退出，不进入服务阶段。
func main() {
	logger, closeLog, err := logging.Init(logging.ConfigFromEnv())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = closeLog()
	}()
	fail := func(message string, attrs ...any) {
		logger.Error(message, attrs...)
		_ = closeLog()
		os.Exit(1)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	buildingCode := os.Getenv("BUILDING_CODE")
	if buildingCode == "" {
		buildingCode = "C5"
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}

	dsn := strings.TrimSpace(os.Getenv("MYSQL_DSN"))
	if dsn == "" {
		fail("MYSQL_DSN 未配置，无法加载楼栋图")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		fail("连接 MySQL 失败", "error", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		fail("获取数据库连接失败", "error", err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		fail("MySQL Ping 失败", "error", err)
	}

	raw, err := repository.NewGraphRepository(db).Load(ctx, buildingCode)
	if err != nil {
		fail("读取图数据失败", "error", err)
	}

	report := graph.Validate(raw, true)
	if !report.Valid {
		for _, issue := range report.Errors {
			logger.Error("图校验失败", "rule", issue.Rule, "table", issue.Table, "id", issue.ID, "message", issue.Message)
		}
		fail("图校验未通过，服务不启动")
	}

	store, err := graph.Build(raw)
	if err != nil {
		fail("构建内存图失败", "error", err)
	}
	stats := store.Stats()
	logger.Info("内存图已加载",
		"building", buildingCode,
		"floors", stats.Floors,
		"rooms", stats.Rooms,
		"facilities", stats.Facilities,
		"nodes", stats.Nodes,
		"edges", stats.Edges,
		"classrooms", stats.Classrooms,
	)

	if os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	catalogService := service.NewCatalogService(store)
	planner := route.NewPlanner(store)

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	var routeCache *cache.RouteCache
	redisClient, redisErr := database.NewRedis(redisAddr, os.Getenv("REDIS_PASSWORD"), envInt("REDIS_DB", 0))
	if redisErr != nil {
		logger.Warn("Redis 不可用，路线缓存关闭", "error", redisErr)
	} else {
		defer redisClient.Close()
		routeCache = cache.NewRouteCache(redisClient, cache.Options{
			BaseTTL: time.Duration(envInt("ROUTE_CACHE_TTL_SECONDS", 900)) * time.Second,
			Jitter:  time.Duration(envInt("ROUTE_CACHE_JITTER_SECONDS", 300)) * time.Second,
		})
		redisCtx, redisCancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := routeCache.EnsureMapVersion(redisCtx); err != nil {
			logger.Warn("初始化 map:version 失败", "error", err)
		}
		redisCancel()
	}

	routeService := service.NewRouteService(catalogService, planner, routeCache)
	nearbyService := service.NewNearbyService(catalogService, planner)

	engine := router.New(router.Dependencies{
		Route:   handler.NewRouteHandler(routeService),
		Nearby:  handler.NewNearbyHandler(nearbyService),
		Catalog: handler.NewCatalogHandler(catalogService),
		Health:  handler.NewHealthHandler(sqlDB, store, version),
	})

	logger.Info("server listening", "port", port)
	if err := engine.Run(":" + port); err != nil {
		fail("server exited", "error", err)
	}
}

func envInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}
