package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"warehouse/internal/application/commands"
	"warehouse/internal/application/queries"
)

// SlotHandler handles HTTP requests related to slots.

type SlotHandler struct {
	reserveSlotsHandler *commands.ReserveSlotsCommandHandler
	findOptimalSlotHandler *queries.FindOptimalSlotQueryHandler
	getShelfStatusHandler *queries.GetShelfStatusQueryHandler
	healthCheckShelfHandler *queries.HealthCheckShelfQueryHandler
}

func NewSlotHandler(
	reserveSlotsHandler *commands.ReserveSlotsCommandHandler,
	findOptimalSlotHandler *queries.FindOptimalSlotQueryHandler,
	getShelfStatusHandler *queries.GetShelfStatusQueryHandler,
	healthCheckShelfHandler *queries.HealthCheckShelfQueryHandler,
) *SlotHandler {
	return &SlotHandler{
		reserveSlotsHandler: reserveSlotsHandler,
		findOptimalSlotHandler: findOptimalSlotHandler,
		getShelfStatusHandler: getShelfStatusHandler,
		healthCheckShelfHandler: healthCheckShelfHandler,
	}
}

func (h *SlotHandler) ReserveSlots(c *gin.Context) {
	var cmd commands.ReserveSlotsCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.reserveSlotsHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Slots reserved successfully"})
}

func (h *SlotHandler) FindOptimalSlot(c *gin.Context) {
	materialType := c.Query("material_type")
	shelfID := c.Query("shelf_id")

	if materialType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "material_type is required"})
		return
	}

	q := queries.FindOptimalSlotQuery{MaterialType: materialType, ShelfID: shelfID}

	slot, err := h.findOptimalSlotHandler.Handle(c.Request.Context(), q)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, slot)
}

func (h *SlotHandler) GetShelfStatus(c *gin.Context) {
	shelfID := c.Param("shelfId")

	q := queries.GetShelfStatusQuery{ShelfID: shelfID}

	status, err := h.getShelfStatusHandler.Handle(c.Request.Context(), q)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *SlotHandler) HealthCheckShelf(c *gin.Context) {
	shelfID := c.Param("shelfId")

	q := queries.HealthCheckShelfQuery{ShelfID: shelfID}

	health, err := h.healthCheckShelfHandler.Handle(c.Request.Context(), q)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, health)
}
