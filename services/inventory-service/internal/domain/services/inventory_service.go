package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"warehouse/internal/domain/entities"
	"warehouse/internal/domain/repositories"
	"warehouse/pkg/errors"
	"warehouse/pkg/utils"
)

// InventoryService provides a high-level interface to the inventory system.
// It is responsible for coordinating the various domain services and repositories
// to perform complex operations.
type InventoryService struct {
	materialRepo    repositories.MaterialRepository
	slotRepo        repositories.SlotRepository
	operationRepo   repositories.OperationRepository
	alertRepo       repositories.AlertRepository
	lockService     *LockService
	eventService    *EventService
	cacheService    *CacheService
	auditService    *AuditService
	alertService    *AlertService
	failedEventRepo repositories.FailedEventRepository
}

// NewInventoryService creates a new instance of the InventoryService.
func NewInventoryService(
	materialRepo repositories.MaterialRepository,
	slotRepo repositories.SlotRepository,
	operationRepo repositories.OperationRepository,
	alertRepo repositories.AlertRepository,
	lockService *LockService,
	eventService *EventService,
	cacheService *CacheService,
	auditService *AuditService,
	alertService *AlertService,
	failedEventRepo repositories.FailedEventRepository,
) *InventoryService {
	return &InventoryService{
		materialRepo:    materialRepo,
		slotRepo:        slotRepo,
		operationRepo:   operationRepo,
		alertRepo:       alertRepo,
		lockService:     lockService,
		eventService:    eventService,
		cacheService:    cacheService,
		auditService:    auditService,
		alertService:    alertService,
		failedEventRepo: failedEventRepo,
	}
}