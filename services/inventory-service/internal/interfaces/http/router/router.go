package router

import (
    "github.com/gin-gonic/gin"
    "warehouse/internal/interfaces/http/handlers"
    "warehouse/internal/interfaces/http/middleware"
)

func SetupInventoryRoutes(r *gin.Engine, inventoryHandler *handlers.InventoryHandler) {
    // apply global middleware
    r.Use(middleware.CORS())
    r.Use(middleware.RequestLogger())
    r.Use(middleware.ErrorHandler())
    
    // group routes under /api/v1
    v1 := r.Group("/api/v1")
    {
        // material operations
        v1.POST("/materials/place", inventoryHandler.PlaceMaterial)
        v1.POST("/materials/remove", inventoryHandler.RemoveMaterial)
        v1.POST("/materials/move", inventoryHandler.MoveMaterial)
        v1.POST("/materials/batch-place", inventoryHandler.BatchPlaceMaterials)
        v1.GET("/materials/search", inventoryHandler.SearchMaterials)
        
        // slot operations
        v1.POST("/slots/reserve", inventoryHandler.ReserveSlots)
        v1.GET("/slots/optimal", inventoryHandler.FindOptimalSlot)
        
        // shelf status info
        v1.GET("/shelves/:shelfId/status", inventoryHandler.GetShelfStatus)
        v1.GET("/shelves/:shelfId/health", inventoryHandler.HealthCheckShelf)
    }
    
    // check health endpoint
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // metrics endpoint for Prometheus
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}