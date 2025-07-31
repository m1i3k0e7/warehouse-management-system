package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"WMS/services/inventory-service/internal/application/queries"
)

// OperationHandler handles HTTP requests related to operations.

type OperationHandler struct {
	getOperationsHandler *queries.GetOperationsQueryHandler
}

func NewOperationHandler(getOperationsHandler *queries.GetOperationsQueryHandler) *OperationHandler {
	return &OperationHandler{getOperationsHandler: getOperationsHandler}
}

func (h *OperationHandler) GetOperations(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	q := queries.GetOperationsQuery{Limit: limit, Offset: offset}

	operations, err := h.getOperationsHandler.Handle(c.Request.Context(), q)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"operations": operations})
}
