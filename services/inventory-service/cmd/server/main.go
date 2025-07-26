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
    // 加載配置
    cfg := config.Load()
    
    // 初始化日誌
    logger.Init(cfg.LogLevel)
    
    // 初始化數據庫
    db, err := database.NewPostgresConnection(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // 初始化 Redis
    redisClient := cache.NewRedisClient(cfg.Redis)
    defer redisClient.Close()
    
    // 初始化 Kafka 事件服務
    eventService, err := services.NewEventService(cfg.Kafka.Brokers, cfg.Kafka.Topic)
    if err != nil {
        log.Fatal("Failed to initialize event service:", err)
    }
    defer eventService.Close()
    
    // 初始化所有服務
    lockService := services.NewLockService(redisClient)
    cacheService := services.NewCacheService(redisClient)
    retryService := services.NewRetryService(3, time.Second*2)
    auditService := services.NewAuditService(eventService)
    alertService := services.NewAlertService(eventService)
    
    // 初始化 repositories
    materialRepo := repositories.NewMaterialRepository(db)
    slotRepo := repositories.NewSlotRepository(db)
    operationRepo := repositories.NewOperationRepository(db)
    alertRepo := repositories.NewAlertRepository(db)
    failedEventRepo := repositories.NewFailedEventRepository(db)
    
    // 初始化庫存服務
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
    
    // 初始化 MQTT 處理器
    mqttHandler := mqtt.NewMQTTHandler(cfg.MQTT.BrokerURL, inventoryService, retryService)
    if err := mqttHandler.Connect(); err != nil {
        log.Fatal("Failed to connect to MQTT broker:", err)
    }
    
    // 初始化 HTTP 路由
    gin.SetMode(cfg.Server.Mode)
    r := router.SetupRouter(db, redisClient, eventService)
    
    // 配置 HTTP 服務器
    srv := &http.Server{
        Addr:         ":" + cfg.Server.Port,
        Handler:      r,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        IdleTimeout:  cfg.Server.IdleTimeout,
    }
    
    // 啟動服務器
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    logger.Info("Inventory service started successfully")
    
    // 優雅關閉
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