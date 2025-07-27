package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"warehouse/internal/application/commands"
	"warehouse/internal/application/queries"
)

// MaterialHandler handles HTTP requests related to materials.

type MaterialHandler struct {
	placeMaterialHandler *commands.PlaceMaterialCommandHandler
	removeMaterialHandler *commands.RemoveMaterialCommandHandler
	moveMaterialHandler *commands.MoveMaterialCommandHandler
	searchMaterialsHandler *queries.SearchMaterialsQueryHandler
}

func NewMaterialHandler(
	placeMaterialHandler *commands.PlaceMaterialCommandHandler,
	removeMaterialHandler *commands.RemoveMaterialCommandHandler,
	moveMaterialHandler *commands.MoveMaterialCommandHandler,
	searchMaterialsHandler *queries.SearchMaterialsQueryHandler,
) *MaterialHandler {
	return &MaterialHandler{
		placeMaterialHandler: placeMaterialHandler,
		removeMaterialHandler: removeMaterialHandler,
		moveMaterialHandler: moveMaterialHandler,
		searchMaterialsHandler: searchMaterialsHandler,
	}
}

func (h *MaterialHandler) PlaceMaterial(c *gin.Context) {
	var cmd commands.PlaceMaterialCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.placeMaterialHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Material placed successfully"})
}

func (h *MaterialHandler) RemoveMaterial(c *gin.Context) {
	var cmd commands.RemoveMaterialCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.removeMaterialHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Material removed successfully"})
}

func (h *MaterialHandler) MoveMaterial(c *gin.Context) {
	var cmd commands.MoveMaterialCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.moveMaterialHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Material moved successfully"})
}

func (h *MaterialHandler) SearchMaterials(c *gin.Context) {
	query := c.Query("q")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	q := queries.SearchMaterialsQuery{Query: query, Limit: limit, Offset: offset}

	materials, err := h.searchMaterialsHandler.Handle(c.Request.Context(), q)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"materials": materials})
}
