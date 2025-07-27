package router

import (
    "github.com/gin-gonic/gin"
    "warehouse/internal/interfaces/http/handlers"
    "warehouse/internal/interfaces/http/middleware"
)

func SetupRoutes(r *gin.Engine, materialHandler *handlers.MaterialHandler, slotHandler *handlers.SlotHandler, operationHandler *handlers.OperationHandler) {
    // apply global middleware
    r.Use(middleware.CORS())
    r.Use(middleware.RequestLogger())
    r.Use(middleware.ErrorHandler())
    
    // group routes under /api/v1
    v1 := r.Group("/api/v1")
    {
        // material operations
        v1.POST("/materials/place", materialHandler.PlaceMaterial)
        v1.POST("/materials/remove", materialHandler.RemoveMaterial)
        v1.POST("/materials/move", materialHandler.MoveMaterial)
        v1.POST("/materials/batch-place", materialHandler.BatchPlaceMaterials)
        v1.GET("/materials/search", materialHandler.SearchMaterials)
        
        // slot operations
        v1.POST("/slots/reserve", slotHandler.ReserveSlots)
        v1.GET("/slots/optimal", slotHandler.FindOptimalSlot)
        
        // shelf status info
        v1.GET("/shelves/:shelfId/status", slotHandler.GetShelfStatus)
        v1.GET("/shelves/:shelfId/health", slotHandler.HealthCheckShelf)

        // operation logs
        v1.GET("/operations", operationHandler.GetOperations)
    }
    
    // check health endpoint
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // metrics endpoint for Prometheus
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}