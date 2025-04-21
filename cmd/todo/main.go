// Package main is the entry point to the application
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/henryhall897/golang-todo-app/database"
	"github.com/henryhall897/golang-todo-app/internal/config"
	"github.com/henryhall897/golang-todo-app/internal/core/logging"
	"github.com/henryhall897/golang-todo-app/internal/middleware"
	"github.com/henryhall897/golang-todo-app/internal/router"
	"github.com/henryhall897/golang-todo-app/internal/server"

	//User packages
	usercache "github.com/henryhall897/golang-todo-app/internal/users/cache"
	userdomains "github.com/henryhall897/golang-todo-app/internal/users/domain"
	userhandlers "github.com/henryhall897/golang-todo-app/internal/users/handler"
	userrepo "github.com/henryhall897/golang-todo-app/internal/users/repository"
	userroutes "github.com/henryhall897/golang-todo-app/internal/users/routes"
	userservices "github.com/henryhall897/golang-todo-app/internal/users/services"

	// Redis wrapper
	rediswrapper "github.com/henryhall897/golang-todo-app/pkg/redis"
)

// Version and BuildDate are populated at build time
var (
	version   string
	buildDate string
)

// AppConfig holds the application configuration
type AppConfig struct {
	Logger   config.LoggingConfig
	Database config.DatabaseConfig
	Server   config.ServerConfig
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load configuration
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logging.InitializeLogger(cfg.Logger.Level, cfg.Logger.Format)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Logger sync failed: %v\n", err)
		}
	}()

	sugarLogger := logger.Sugar()

	// Graceful shutdown handling
	defer func() {
		if r := recover(); r != nil {
			sugarLogger.Fatal("Application panicked", "error", r)
		}
		stop()
	}()

	// Run the application
	err = run(ctx, sugarLogger, cfg)
	if err != nil {
		sugarLogger.Errorw("Application error", "error", err)
		os.Exit(1)
	}

	sugarLogger.Info("Application shut down successfully")
}

func run(ctx context.Context, logger *zap.SugaredLogger, cfg *config.AppConfig) error {
	// Log version details
	if version == "" {
		version = "dev"
	}
	if buildDate == "" {
		buildDate = time.Now().Format(time.RFC3339) // Fallback to current time
	}
	logger.Infow("Version Information", "version", version, "buildDate", buildDate)

	// Initialize database connection
	logger.Info("Initializing database connection")
	pool, err := database.InitializeDatabasePool(logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Run database migrations
	logger.Info("Running database migrations")
	if err := database.ApplyMigrations(ctx, cfg.Database.DatabaseURL); err != nil {
		return fmt.Errorf("migration error: %w", err)
	}
	logger.Info("Database migrations completed successfully")

	// Initialize Redis Client
	logger.Info("Initializing Redis connection")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password, // Leave empty if no password
		DB:       cfg.Redis.DB,       // Use default DB
	})

	// Check Redis connectivity
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	logger.Info("Redis connection established successfully")

	// Initialize Redis cache
	// Generic redis cache
	genericCache := rediswrapper.NewJSONCache(redisClient, userdomains.RedisPrefix, logger)
	//User specific redis cache
	userCache := usercache.NewRedisUser(genericCache)

	// Initialize stores
	userStore := userrepo.New(pool)

	// Initialize services
	userService := userservices.New(userStore, userCache, logger)

	// Initialize HTTP handlers
	userHandler := userhandlers.New(userService, logger)

	// Register route functions
	routeFuncs := []router.RouteRegisterFunc{
		userroutes.RegisterRoutes,
	}

	// Initialize the router
	rt := router.NewRouter(routeFuncs, []userhandlers.Handler{*userHandler})

	// TODO - Add more route modules here (e.g., tasks, lists)

	// Apply CORS middleware to router
	corsWrappedHandler := middleware.CORS(cfg.Server.CorsOrigin)(rt.LimitedHandler)

	// Start the HTTP server
	srv := server.NewHTTPServer(&config.ServerConfig{
		BindAddress: cfg.Server.BindAddress,
		Port:        cfg.Server.Port,
		Logger:      logger,
	})
	return srv.Serve(ctx, corsWrappedHandler)
}
