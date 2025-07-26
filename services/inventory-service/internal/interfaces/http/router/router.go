package router

import (
    "github.com/gin-gonic/gin"
    "warehouse/internal/interfaces/http/handlers"
    "warehouse/internal/interfaces/http/middleware"
)

func SetupInventoryRoutes(r *gin.Engine, inventoryHandler *handlers.InventoryHandler) {
    // 添加中介軟體
    r.Use(middleware.CORS())
    r.Use(middleware.RequestLogger())
    r.Use(middleware.ErrorHandler())
    
    // API 版本分組
    v1 := r.Group("/api/v1")
    {
        // 材料操作
        v1.POST("/materials/place", inventoryHandler.PlaceMaterial)
        v1.POST("/materials/remove", inventoryHandler.RemoveMaterial)
        v1.POST("/materials/move", inventoryHandler.MoveMaterial)
        v1.POST("/materials/batch-place", inventoryHandler.BatchPlaceMaterials)
        v1.GET("/materials/search", inventoryHandler.SearchMaterials)
        
        // 格子操作
        v1.POST("/slots/reserve", inventoryHandler.ReserveSlots)
        v1.GET("/slots/optimal", inventoryHandler.FindOptimalSlot)
        
        // 料架狀態
        v1.GET("/shelves/:shelfId/status", inventoryHandler.GetShelfStatus)
        v1.GET("/shelves/:shelfId/health", inventoryHandler.HealthCheckShelf)
    }
    
    // 健康檢查端點
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // 指標端點
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}