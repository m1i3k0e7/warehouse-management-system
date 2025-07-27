package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "warehouse/internal/domain/services"
    "warehouse/pkg/logger"
)

type InventoryHandler struct {
    inventoryService *services.InventoryService
}

func NewInventoryHandler(inventoryService *services.InventoryService) *InventoryHandler {
    return &InventoryHandler{
        inventoryService: inventoryService,
    }
}

func (h *InventoryHandler) PlaceMaterial(c *gin.Context) {
    var cmd services.PlaceMaterialCommand
    if err := c.ShouldBindJSON(&cmd); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.inventoryService.PlaceMaterial(c.Request.Context(), cmd); err != nil {
        logger.Error("Failed to place material", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Material placed successfully"})
}

func (h *InventoryHandler) RemoveMaterial(c *gin.Context) {
    var cmd services.RemoveMaterialCommand
    if err := c.ShouldBindJSON(&cmd); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.inventoryService.RemoveMaterial(c.Request.Context(), cmd); err != nil {
        logger.Error("Failed to remove material", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Material removed successfully"})
}

func (h *InventoryHandler) MoveMaterial(c *gin.Context) {
    var cmd services.MoveMaterialCommand
    if err := c.ShouldBindJSON(&cmd); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.inventoryService.MoveMaterial(c.Request.Context(), cmd); err != nil {
        logger.Error("Failed to move material", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Material moved successfully"})
}

func (h *InventoryHandler) GetShelfStatus(c *gin.Context) {
    shelfID := c.Param("shelfId")
    
    status, err := h.inventoryService.GetShelfStatus(c.Request.Context(), shelfID)
    if err != nil {
        logger.Error("Failed to get shelf status", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, status)
}

func (h *InventoryHandler) FindOptimalSlot(c *gin.Context) {
    materialType := c.Query("material_type")
    shelfID := c.Query("shelf_id")
    
    if materialType == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "material_type is required"})
        return
    }
    
    slot, err := h.inventoryService.FindOptimalSlot(c.Request.Context(), materialType, shelfID)
    if err != nil {
        logger.Error("Failed to find optimal slot", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, slot)
}

func (h *InventoryHandler) ReserveSlots(c *gin.Context) {
    var cmd services.ReserveSlotsCommand
    if err := c.ShouldBindJSON(&cmd); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.inventoryService.ReserveSlots(c.Request.Context(), cmd); err != nil {
        logger.Error("Failed to reserve slots", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Slots reserved successfully"})
}

func (h *InventoryHandler) HealthCheckShelf(c *gin.Context) {
    shelfID := c.Param("shelfId")
    
    health, err := h.inventoryService.HealthCheckShelf(c.Request.Context(), shelfID)
    if err != nil {
        logger.Error("Failed to perform health check", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, health)
}

func (h *InventoryHandler) BatchPlaceMaterials(c *gin.Context) {
    var commands []services.PlaceMaterialCommand
    if err := c.ShouldBindJSON(&commands); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.inventoryService.BatchPlaceMaterials(c.Request.Context(), commands); err != nil {
        logger.Error("Failed to batch place materials", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Materials placed successfully"})
}

func (h *InventoryHandler) SearchMaterials(c *gin.Context) {
    query := c.Query("q")
    limitStr := c.DefaultQuery("limit", "20")
    offsetStr := c.DefaultQuery("offset", "0")
    
    limit, _ := strconv.Atoi(limitStr)
    offset, _ := strconv.Atoi(offsetStr)
    
    materials, err := h.inventoryService.SearchMaterials(c.Request.Context(), query, limit, offset)
    if err != nil {
        logger.Error("Failed to search materials", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"materials": materials})
}