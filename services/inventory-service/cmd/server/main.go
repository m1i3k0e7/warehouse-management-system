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
    
    // 初始化 Kafka
    kafkaProducer := messaging.NewKafkaProducer(cfg.Kafka)
    defer kafkaProducer.Close()
    
    // 初始化 Gin 路由
    gin.SetMode(cfg.Server.Mode)
    r := router.SetupRouter(db, redisClient, kafkaProducer)
    
    // 配置 HTTP 服務器
    srv := &http.Server{
        Addr:         ":" + cfg.Server.Port,
        Handler:      r,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        IdleTimeout:  cfg.Server.IdleTimeout,
    }
    
    // 優雅關閉
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    // 等待中斷信號
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exited")
}