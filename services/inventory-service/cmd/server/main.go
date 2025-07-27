package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "warehouse/internal/config"
    "warehouse/internal/infrastructure/database"
    "warehouse/internal/infrastructure/cache"
    "warehouse/internal/infrastructure/messaging"
    "warehouse/internal/interfaces/http/router"
    "warehouse/internal/interfaces/mqtt"
    "warehouse/internal/domain/services"
    "warehouse/pkg/logger"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize logger
    logger.Init(cfg.LogLevel)
    
    // Initialize database connection
    db, err := database.NewPostgresConnection(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Initialize Redis client
    redisClient := cache.NewRedisClient(cfg.Redis)
    defer redisClient.Close()
    
    // Initialize Kafka event service
    eventService, err := services.NewEventService(cfg.Kafka.Brokers, cfg.Kafka.Topic)
    if err != nil {
        log.Fatal("Failed to initialize event service:", err)
    }
    defer eventService.Close()
    
    // Initialize all other services
    lockService := services.NewLockService(redisClient)
    cacheService := services.NewCacheService(redisClient)
    retryService := services.NewRetryService(cfg.ServiceConfig.RetryCount, cfg.ServiceConfig.RetryDelay)
    auditService := services.NewAuditService(eventService)
    alertService := services.NewAlertService(eventService)
    
    // Initialize repositories
    materialRepo := repositories.NewMaterialRepository(db)
    slotRepo := repositories.NewSlotRepository(db)
    operationRepo := repositories.NewOperationRepository(db)
    alertRepo := repositories.NewAlertRepository(db)
    failedEventRepo := repositories.NewFailedEventRepository(db)
    
    // Initialize inventory service
    inventoryService := services.NewInventoryService(
        materialRepo,
        slotRepo,
        operationRepo,
        alertRepo,
        lockService,
        eventService,
        cacheService,
        auditService,
        alertService,
        failedEventRepo,
    )
    
    // Initialize MQTT handler
    mqttHandler := mqtt.NewMQTTHandler(cfg.MQTT.BrokerURL, inventoryService, retryService)
    if err := mqttHandler.Connect(); err != nil {
        log.Fatal("Failed to connect to MQTT broker:", err)
    }
    
    // Initialize http router
    gin.SetMode(cfg.Server.Mode)
    r := router.SetupRouter(db, redisClient, eventService)
    
    // configure http server
    srv := &http.Server{
        Addr:         ":" + cfg.Server.Port,
        Handler:      r,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        IdleTimeout:  cfg.Server.IdleTimeout,
    }
    
    // start the server
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    logger.Info("Inventory service started successfully")
    
    // handle shutdown signals
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    logger.Info("Server exited")
}