package middleware

import (
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "time"
    "warehouse/internal/config"
)

func CORS() gin.HandlerFunc {
    cfg := config.Load()
    return cors.New(cors.Config{
        AllowOrigins:     []string{cfg.ServiceConfig.AllowedOrigin},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}