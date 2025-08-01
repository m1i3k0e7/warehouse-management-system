package services

import (
	"context"
	"fmt"
	"time"

	
	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/internal/domain/repositories"
	"WMS/services/inventory-service/pkg/errors"
	"WMS/services/inventory-service/pkg/utils/logger"

	"github.com/google/uuid"
)

type PlaceMaterialParams struct {
	MaterialBarcode string
	SlotID          string
	OperatorID      string
}

type RemoveMaterialParams struct {
	SlotID     string
	OperatorID string
	Reason     string
}

type MoveMaterialParams struct {
	FromSlotID string
	ToSlotID   string
	OperatorID string
	Reason     string
}

type ReserveSlotsParams struct {
	SlotIDs    []string
	OperatorID string
	Duration   int
	Purpose    string
}

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
	retryService 	*RetryService
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
	retryService *RetryService,
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
		retryService: 	 retryService,
		failedEventRepo: failedEventRepo,
	}
}

func (s *InventoryService) PlaceMaterial(ctx context.Context, param PlaceMaterialParams) error {
	// validate command parameters
	if err := s.validatePlaceMaterialParams(param); err != nil {
		return errors.NewValidationError("invalid command", err)
	}

	// acquire lock on the shelf
	slot, err := s.slotRepo.GetByID(ctx, param.SlotID)
	if err != nil {
		return errors.NewNotFoundError("slot not found", err)
	}

	lockKey := fmt.Sprintf("shelf:%s", slot.ShelfID)
	unlock, err := s.lockService.AcquireLock(ctx, lockKey, 30*time.Second)
	if err != nil {
		return errors.NewConflictError("shelf is locked", err)
	}
	defer unlock()

	// validate preconditions for placing material
	if err := s.validatePlacementPreconditions(ctx, param); err != nil {
		return err
	}

	// execute the placement operation
	if err := s.retryService.ExecuteWithRetry(ctx, s.executePlaceMaterial, param, entities.OperationTypeMove); err != nil {
		// log the failed operation
		s.auditService.LogFailedOperation(ctx, "place_material", param, err)
		s.SaveFailedEventToDLQ(ctx, EventTypeMaterialPlaced, EventTypeMaterialPlaced, param, err)
		return err
	}

	// check for anomalies
	// s.checkForAnomalies(ctx, operation, params.SensorData)

	// log the successful operation
	// s.auditService.LogSuccessfulOperation(ctx, operation)

	return nil
}

func (s *InventoryService) RemoveMaterial(ctx context.Context, param RemoveMaterialParams) error {
	slot, err := s.slotRepo.GetByID(ctx, param.SlotID)
	if err != nil {
		return errors.NewNotFoundError("slot not found", err)
	}

	if slot.Status != entities.SlotStatusOccupied {
		return errors.NewConflictError("slot is empty", nil)
	}

	lockKey := fmt.Sprintf("shelf:%s", slot.ShelfID)
	unlock, err := s.lockService.AcquireLock(ctx, lockKey, 30*time.Second)
	if err != nil {
		return errors.NewConflictError("shelf is locked", err)
	}
	defer unlock()

	if err := s.retryService.ExecuteWithRetry(ctx, s.executeRemoveMaterial, param, entities.OperationTypeMove); err != nil {
		// log the failed operation
		s.auditService.LogFailedOperation(ctx, "remove_material", param, err)
		s.SaveFailedEventToDLQ(ctx, EventTypeMaterialRemoved, EventTypeMaterialRemoved, param, err)
		return err
	}

	return nil
}

func (s *InventoryService) MoveMaterial(ctx context.Context, params MoveMaterialParams) error {
	// get the source and target slots
	fromSlot, err := s.slotRepo.GetByID(ctx, params.FromSlotID)
	if err != nil {
		return errors.NewNotFoundError("source slot not found", err)
	}

	toSlot, err := s.slotRepo.GetByID(ctx, params.ToSlotID)
	if err != nil {
		return errors.NewNotFoundError("target slot not found", err)
	}

	// validate the source and target slots status
	if fromSlot.Status != entities.SlotStatusOccupied {
		return errors.NewConflictError("source slot is empty", nil)
	}
	if toSlot.Status != entities.SlotStatusEmpty {
		return errors.NewConflictError("target slot is not empty", nil)
	}

	// acquire locks on both source and target shelves
	locks := s.acquireMultipleShelfLocks(ctx, []string{fromSlot.ShelfID, toSlot.ShelfID})
	defer s.releaseMultipleLocks(locks)

	if err := s.retryService.ExecuteWithRetry(ctx, s.executeMoveMaterial, params, entities.OperationTypeMove); err != nil {
		// log the failed operation
		s.auditService.LogFailedOperation(ctx, "move_material", params, err)
		s.SaveFailedEventToDLQ(ctx, EventTypeMaterialMoved, EventTypeMaterialMoved, params, err)
		return err
	}

	return nil
}

func (s *InventoryService) ReserveSlots(ctx context.Context, param ReserveSlotsParams) error {
	// acquire locks on all shelves involved in the reservation
	shelfIDs := make([]string, 0)
	slotShelfMap := make(map[string]string)
	for _, slotID := range param.SlotIDs {
		slot, err := s.slotRepo.GetByID(ctx, slotID)
		if err != nil {
			return errors.NewNotFoundError(fmt.Sprintf("slot %s not found", slotID), err)
		}

		slotShelfMap[slotID] = slot.ShelfID
		found := false
		for _, shelfID := range shelfIDs {
			if shelfID == slot.ShelfID {
				found = true
				break
			}
		}

		if !found {
			shelfIDs = append(shelfIDs, slot.ShelfID)
		}
	}

	locks := s.acquireMultipleShelfLocks(ctx, shelfIDs)
	defer s.releaseMultipleLocks(locks)

	if err := s.retryService.ExecuteWithRetry(ctx, s.executeReserveSlots, param, slotShelfMap); err != nil {
		// log the failed operation
		s.auditService.LogFailedOperation(ctx, "reserve_slots", param, err)
		s.SaveFailedEventToDLQ(ctx, EventTypeSlotsReserved, EventTypeSlotsReserved, param, err)
		return err
	}

	return nil
}

func (s *InventoryService) FindOptimalSlot(ctx context.Context, materialType string, shelfID string) (*entities.Slot, error) {
	slots, err := s.slotRepo.GetEmptySlotsByShelf(ctx, shelfID)
	if err != nil {
		return nil, err
	}

	if len(slots) == 0 {
		return nil, errors.NewNotFoundError("no empty slots available", nil)
	}

	return s.selectBestSlot(slots, materialType)
}

func (s *InventoryService) BatchPlaceMaterials(ctx context.Context, params []PlaceMaterialParams) error {
	// group commands by shelf
	shelfGroups, err := s.groupCommandsByShelf(params)
	if err != nil {
		return errors.NewInternalError("failed to group commands by shelf", err)
	}
	
	for shelfID, shelfParams := range shelfGroups {
		lockKey := fmt.Sprintf("shelf:%s", shelfID)
		unlock, err := s.lockService.AcquireLock(ctx, lockKey, 60*time.Second)
		if err != nil {
			return errors.NewConflictError(fmt.Sprintf("failed to lock shelf %s", shelfID), err)
		}
		defer unlock()

		err = s.executeBatchPlacement(ctx, shelfParams)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *InventoryService) HealthCheckShelf(ctx context.Context, shelfID string) (*entities.ShelfHealth, error) {
	slots, err := s.slotRepo.GetByShelfID(ctx, shelfID)
	if err != nil {
		return nil, err
	}

	health := &entities.ShelfHealth{
		ShelfID:          shelfID,
		TotalSlots:       len(slots),
		HealthySlots:     0,
		ErrorSlots:       0,
		MaintenanceSlots: 0,
		LastCheckTime:    time.Now(),
	}

	for _, slot := range slots {
		switch slot.Status {
		case entities.SlotStatusEmpty, entities.SlotStatusOccupied:
			health.HealthySlots++
		case entities.SlotStatusMaintenance:
			health.MaintenanceSlots++
		default:
			health.ErrorSlots++
		}
	}

	health.HealthScore = float64(health.HealthySlots) / float64(health.TotalSlots) * 100

	// send health alert if score is below threshold
	if health.HealthScore < 95.0 {
		s.alertService.SendShelfHealthAlert(ctx, health)
	}

	return health, nil
}

func (s *InventoryService) HandleSlotError(ctx context.Context, slotID string, errorType string) error {
	slot, err := s.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		return err
	}

	// log the error
	alert := &entities.Alert{
		ID:        generateUUID(),
		Type:      "slot_error",
		ShelfID:   slot.ShelfID,
		SlotID:    slotID,
		Message:   fmt.Sprintf("Slot error: %s", errorType),
		Severity:  "high",
		CreatedAt: time.Now(),
		Status:    "active",
	}
	if err := s.alertRepo.Create(ctx, alert); err != nil {
		logger.Error("Failed to create alert", err)
	}

	// handle the error based on its type
	switch errorType {
	case "sensor_error":
		return s.markSlotForMaintenance(ctx, slotID, "sensor malfunction")
	case "weight_mismatch":
		return s.triggerManualVerification(ctx, slotID)
	default:
		return s.markSlotForInvestigation(ctx, slotID, errorType)
	}
}

func (s *InventoryService) UpdateShelfStatus(ctx context.Context, shelfID string, status string) error {
	// update the shelf status in the cache
	cacheKey := fmt.Sprintf("shelf_status:%s", shelfID)
	statusData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	return s.cacheService.Set(ctx, cacheKey, statusData, time.Hour)
}

func (s *InventoryService) GetShelfStatus(ctx context.Context, shelfID string) (*entities.ShelfStatus, error) {
	// attempt to get the shelf status from the cache
	if status, err := s.cacheService.GetShelfStatus(ctx, shelfID); err == nil && status != nil {
		return status, nil
	}

	// retrieve the shelf slots from the database
	slots, err := s.slotRepo.GetByShelfID(ctx, shelfID)
	if err != nil {
		return nil, errors.NewInternalError("failed to get shelf slots", err)
	}

	status := &entities.ShelfStatus{
		ShelfID:       shelfID,
		TotalSlots:    len(slots),
		EmptySlots:    0,
		OccupiedSlots: 0,
		Slots:         make([]entities.Slot, len(slots)),
		UpdatedAt:     time.Now(),
	}

	for i, slot := range slots {
		status.Slots[i] = *slot
		switch slot.Status {
		case entities.SlotStatusEmpty:
			status.EmptySlots++
		case entities.SlotStatusOccupied:
			status.OccupiedSlots++
		}
	}

	// update the cache with the current shelf status
	s.cacheService.SetShelfStatus(ctx, status)

	return status, nil
}

func (s *InventoryService) validatePlaceMaterialParams(params PlaceMaterialParams) error {
	if params.MaterialBarcode == "" || params.SlotID == "" || params.OperatorID == "" {
		return fmt.Errorf("material barcode, slot ID, and operator ID are required")
	}

	return nil
}

func (s *InventoryService) validatePlacementPreconditions(ctx context.Context, params PlaceMaterialParams) error {
	slot, err := s.slotRepo.GetByID(ctx, params.SlotID)
	if err != nil {
		return errors.NewNotFoundError("slot not found", err)
	}
	if slot.Status != entities.SlotStatusEmpty {
		return errors.NewConflictError("slot is not available", nil)
	}

	material, err := s.materialRepo.GetByBarcode(ctx, params.MaterialBarcode)
	if err != nil {
		return errors.NewNotFoundError("material not found", err)
	}
	if material.Status != entities.MaterialStatusAvailable {
		return errors.NewConflictError("material is not available", nil)
	}

	return nil
}

func (s *InventoryService) executePlaceMaterial(ctx context.Context, params PlaceMaterialParams) error {
	tx, err := s.slotRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	slot, _ := s.slotRepo.GetByID(ctx, params.SlotID)
	material, _ := s.materialRepo.GetByBarcode(ctx, params.MaterialBarcode)
	slot.Status = entities.SlotStatusOccupied
	slot.MaterialID = &material.ID
	slot.UpdatedAt = time.Now()
	slot.Version++
	if err := s.slotRepo.UpdateWithTx(ctx, tx, slot); err != nil {
		return errors.NewConflictError("failed to update slot", err)
	}

	material.Status = entities.MaterialStatusInUse
	material.UpdatedAt = time.Now()
	if err := s.materialRepo.UpdateWithTx(ctx, tx, material); err != nil {
		return errors.NewInternalError("failed to update material", err)
	}

	operation := &entities.Operation{
		ID:         generateUUID(),
		Type:       entities.OperationTypePlacement,
		MaterialID: material.ID,
		SlotID:     params.SlotID,
		OperatorID: params.OperatorID,
		ShelfID:    slot.ShelfID,
		Timestamp:  time.Now(),
		Status:     entities.OperationStatusPendingPhysicalConfirmation,
	}
	if err := s.operationRepo.CreateWithTx(ctx, tx, operation); err != nil {
		return errors.NewInternalError("failed to record operation", err)
	}

	if err := tx.Commit(); err != nil {
		return errors.NewInternalError("failed to commit transaction", err.Error)
	}

	// Publish event to request physical placement
	s.publishPhysicalPlacementRequestedEvent(ctx, operation)
	return nil
}

func (s *InventoryService) executeRemoveMaterial(ctx context.Context, param RemoveMaterialParams) error {
	tx, err := s.slotRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	slot, err := s.slotRepo.GetByID(ctx, param.SlotID)
	if err != nil {
		return errors.NewNotFoundError("slot not found", err)
	}

	material, err := s.materialRepo.GetByID(ctx, *slot.MaterialID)
	if err != nil {
		return errors.NewNotFoundError("material not found", err)
	}

	slot.Status = entities.SlotStatusRemovalPending
	slot.UpdatedAt = time.Now()
	slot.Version++
	if err := s.slotRepo.UpdateWithTx(ctx, tx, slot); err != nil {
		return errors.NewConflictError("failed to update slot", err)
	}

	operation := &entities.Operation{
		ID:         generateUUID(),
		Type:       entities.OperationTypeRemoval,
		MaterialID: material.ID,
		SlotID:     param.SlotID,
		OperatorID: param.OperatorID,
		ShelfID:    slot.ShelfID,
		Timestamp:  time.Now(),
		// Status:     entities.OperationStatusCompleted,
		Status: entities.OperationStatusPendingRemovalConfirmation,
	}
	if err := s.operationRepo.CreateWithTx(ctx, tx, operation); err != nil {
		return errors.NewInternalError("failed to record operation", err)
	}

	if err := tx.Commit(); err != nil {
		return errors.NewInternalError("failed to commit transaction", err.Error)
	}

	s.publishMaterialRemovedEvent(ctx, operation)

	return nil
}

func (s *InventoryService) executeMoveMaterial(ctx context.Context, param MoveMaterialParams) error {
	tx, err := s.slotRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	fromSlot, err := s.slotRepo.GetByID(ctx, param.FromSlotID)
	if err != nil {
		return errors.NewNotFoundError("source slot not found", err)
	}

	toSlot, err := s.slotRepo.GetByID(ctx, param.ToSlotID)
	if err != nil {
		return errors.NewNotFoundError("target slot not found", err)
	}

	material, err := s.materialRepo.GetByID(ctx, *fromSlot.MaterialID)
	if err != nil {
		return errors.NewNotFoundError("material not found", err)
	}

	// Update fromSlot
	fromSlot.Status = entities.SlotStatusEmpty
	fromSlot.MaterialID = nil
	fromSlot.UpdatedAt = time.Now()
	fromSlot.Version++
	if err := s.slotRepo.UpdateWithTx(ctx, tx, fromSlot); err != nil {
		return errors.NewConflictError("failed to update from_slot", err)
	}

	// Update toSlot
	toSlot.Status = entities.SlotStatusOccupied
	toSlot.MaterialID = &material.ID
	toSlot.UpdatedAt = time.Now()
	toSlot.Version++
	if err := s.slotRepo.UpdateWithTx(ctx, tx, toSlot); err != nil {
		return errors.NewConflictError("failed to update to_slot", err)
	}

	operation := &entities.Operation{
		ID:         generateUUID(),
		Type:       entities.OperationTypeMove,
		MaterialID: material.ID,
		SlotID:     param.ToSlotID,
		OperatorID: param.OperatorID,
		ShelfID:    toSlot.ShelfID,
		Timestamp:  time.Now(),
		Status:     entities.OperationStatusCompleted,
	}
	if err := s.operationRepo.CreateWithTx(ctx, tx, operation); err != nil {
		return errors.NewInternalError("failed to record operation", err)
	}

	if err := tx.Commit(); err != nil {
		return errors.NewInternalError("failed to commit transaction", err.Error)
	}

	s.publishMaterialMovedEvent(ctx, operation, param.FromSlotID)

	return nil
}

func (s *InventoryService) executeReserveSlots(ctx context.Context, params ReserveSlotsParams, slotShelfMap map[string]string) error {
	tx, err := s.slotRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	for _, slotID := range params.SlotIDs {
		slot, err := s.slotRepo.GetByID(ctx, slotID)
		if err != nil {
			return errors.NewNotFoundError(fmt.Sprintf("slot %s not found", slotID), err)
		}
		if slot.Status != entities.SlotStatusEmpty {
			return errors.NewConflictError(fmt.Sprintf("slot %s is not empty", slotID), nil)
		}
		slot.Status = entities.SlotStatusReserved
		slot.UpdatedAt = time.Now()
		slot.Version++
		if err := s.slotRepo.UpdateWithTx(ctx, tx, slot); err != nil {
			return errors.NewConflictError(fmt.Sprintf("failed to reserve slot %s", slotID), err)
		}
	}
	if err := tx.Commit(); err != nil {
		return errors.NewInternalError("failed to commit transaction", err.Error)
	}
	return nil
}

func (s *InventoryService) selectBestSlot(slots []*entities.Slot, materialType string) (*entities.Slot, error) {
	// This is a simple implementation. A more advanced version could consider
	// proximity to other materials of the same type, operator ergonomics, etc.
	for _, slot := range slots {
		if slot.Status == entities.SlotStatusEmpty {
			if slot.IsSuitableForMaterialType(materialType) {
				return slot, nil
			}
		}
	}

	return nil, fmt.Errorf("no empty slots")
}

func (s *InventoryService) groupCommandsByShelf(params []PlaceMaterialParams) (map[string][]PlaceMaterialParams, error) {
	shelfGroups := make(map[string][]PlaceMaterialParams)
	for _, p := range params {
		slot, err := s.slotRepo.GetByID(context.Background(), p.SlotID)
		if err != nil {
			return nil, errors.NewNotFoundError(fmt.Sprintf("slot %s not found", p.SlotID), err)
		}
		shelfGroups[slot.ShelfID] = append(shelfGroups[slot.ShelfID], p)
	}
	return shelfGroups, nil
}

func (s *InventoryService) executeBatchPlacement(ctx context.Context, params []PlaceMaterialParams) error {
	tx, err := s.slotRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	
	for _, p := range params {
		err := s.retryService.ExecuteWithRetry(ctx, s.executePlaceMaterial, p, entities.OperationTypeMove)
		if err != nil {
			return err // Or collect errors and return them all
		}
	}

	return tx.Commit().Error
}

func (s *InventoryService) markSlotForMaintenance(ctx context.Context, slotID, reason string) error {
	slot, err := s.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		return err
	}

	slot.Status = entities.SlotStatusMaintenance
	slot.UpdatedAt = time.Now()
	slot.Version++
	if err := s.slotRepo.Update(ctx, slot); err != nil {
		return err
	}

	s.publishSystemAlertEvent(ctx, "slot_maintenance", "warning", fmt.Sprintf("Slot %s marked for maintenance: %s", slotID, reason), map[string]interface{}{
		"slot_id": slotID,
		"reason":  reason,
	})

	return nil
}

func (s *InventoryService) triggerManualVerification(ctx context.Context, slotID string) error {
	s.publishSystemAlertEvent(ctx, "manual_verification_required", "high", fmt.Sprintf("Manual verification required for slot %s", slotID), map[string]interface{}{
		"slot_id": slotID,
	})

	return nil
}

func (s *InventoryService) markSlotForInvestigation(ctx context.Context, slotID, reason string) error {
	s.publishSystemAlertEvent(ctx, "slot_investigation", "medium", fmt.Sprintf("Slot %s requires investigation: %s", slotID, reason), map[string]interface{}{
		"slot_id": slotID,
		"reason":  reason,
	})

	return nil
}

func (s *InventoryService) publishSystemAlertEvent(ctx context.Context, alertType entities.AlertType, severity entities.AlertSeverity, message string, details map[string]interface{}) {
	event := &entities.SystemAlertEvent{
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
		Details:   details,
	}

	if err := s.eventService.PublishEvent(ctx, "system_alert", event); err != nil {
		logger.Error("Failed to publish system alert event", err)
	}
}

func (s *InventoryService) acquireMultipleShelfLocks(ctx context.Context, shelfIDs []string) []func() {
	unlockFuncs := make([]func(), 0)
	for _, shelfID := range shelfIDs {
		lockKey := fmt.Sprintf("shelf:%s", shelfID)
		unlock, err := s.lockService.AcquireLock(ctx, lockKey, 30*time.Second)
		if err == nil {
			unlockFuncs = append(unlockFuncs, unlock)
		}
	}

	return unlockFuncs
}

func (s *InventoryService) releaseMultipleLocks(unlockFuncs []func()) {
	for _, unlock := range unlockFuncs {
		unlock()
	}
}

func (s *InventoryService) SearchMaterials(ctx context.Context, query string, limit, offset int) ([]*entities.Material, error) {
	// Search materials by barcode or name
	materials, err := s.materialRepo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, errors.NewInternalError("failed to search materials", err)
	}

	if len(materials) == 0 {
		return nil, errors.NewNotFoundError("no materials found", nil)
	}

	return materials, nil
}

func generateUUID() string {
	return uuid.New().String()
}

// ConfirmPhysicalPlacement confirms that a physical placement operation has been completed.
func (s *InventoryService) ConfirmPhysicalPlacement(ctx context.Context, operationID string) error {
	operation, err := s.operationRepo.GetByID(ctx, operationID)
	if err != nil {
		return errors.NewNotFoundError("operation not found", err)
	}

	if operation.Status != entities.OperationStatusPendingPhysicalConfirmation {
		return errors.NewConflictError(fmt.Sprintf("operation %s is not in pending physical confirmation status", operationID), nil)
	}

	tx, err := s.operationRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	operation.Status = entities.OperationStatusCompleted
	operation.Timestamp = time.Now()
	if err := s.operationRepo.UpdateWithTx(ctx, tx, operation); err != nil {
		return errors.NewInternalError("failed to update operation status", err)
	}

	if err := tx.Commit(); err != nil {
		return errors.NewInternalError("failed to commit transaction", err.Error)
	}

	// Publish material placed confirmed event (now that physical placement is confirmed)
	s.publishPhysicalPlacementConfirmedEvent(ctx, operation)

	return nil
}

func (s *InventoryService) ConfirmPhysicalRemoval(ctx context.Context, operationID string, unplanned bool) error {
	operation, err := s.operationRepo.GetByID(ctx, operationID)
	if err != nil {
		return errors.NewNotFoundError("operation not found", err)
	}

	if operation.Status != entities.OperationStatusPendingRemovalConfirmation {
		return errors.NewConflictError(fmt.Sprintf("operation %s is not in pending physical removal status", operationID), nil)
	}

	tx, err := s.operationRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	slot, err := s.slotRepo.GetByID(ctx, operation.SlotID)
	if err != nil {
		return errors.NewNotFoundError("slot not found for confirmation", err)
	}
	slot.Status = entities.SlotStatusEmpty
	slot.MaterialID = nil
	slot.UpdatedAt = time.Now()
	slot.Version++
	if err := s.slotRepo.UpdateWithTx(ctx, tx, slot); err != nil {
		return errors.NewInternalError("failed to update slot status", err)
	}

	material, err := s.materialRepo.GetByID(ctx, operation.MaterialID)
	if err != nil {
		return errors.NewNotFoundError("material not found for confirmation", err)
	}
	material.Status = entities.MaterialStatusAvailable
	material.UpdatedAt = time.Now()
	if err := s.materialRepo.UpdateWithTx(ctx, tx, material); err != nil {
		return errors.NewInternalError("failed to update material", err)
	}

	operation.Status = entities.OperationStatusCompleted
	operation.Timestamp = time.Now()
	if err := s.operationRepo.UpdateWithTx(ctx, tx, operation); err != nil {
		return errors.NewInternalError("failed to update operation status", err)
	}

	if err := tx.Commit(); err != nil {
		return errors.NewInternalError("failed to commit transaction", err.Error)
	}

	s.publishPhysicalRemovalConfirmedEvent(ctx, operation)
	
	return nil
}

// HandlePhysicalPlacementTimeout handles a physical placement operation that has timed out.
func (s *InventoryService) HandlePhysicalPlacementTimeout(ctx context.Context, operationID string) error {
	operation, err := s.operationRepo.GetByID(ctx, operationID)
	if err != nil {
		return errors.NewNotFoundError("operation not found", err)
	}

	if operation.Status != entities.OperationStatusPendingPhysicalConfirmation {
		return errors.NewConflictError(fmt.Sprintf("operation %s is not in pending physical confirmation status", operationID), nil)
	}

	tx, err := s.operationRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Rollback slot status
	slot, err := s.slotRepo.GetByID(ctx, operation.SlotID)
	if err != nil {
		return errors.NewNotFoundError("slot not found for rollback", err)
	}
	slot.Status = entities.SlotStatusEmpty
	slot.MaterialID = nil
	slot.UpdatedAt = time.Now()
	slot.Version++
	if err := s.slotRepo.UpdateWithTx(ctx, tx, slot); err != nil {
		return errors.NewInternalError("failed to rollback slot status", err)
	}

	// Rollback material status (if applicable, e.g., if it was marked as in_use)
	material, err := s.materialRepo.GetByID(ctx, operation.MaterialID)
	if err != nil {
		return errors.NewNotFoundError("material not found for rollback", err)
	}
	material.Status = entities.MaterialStatusAvailable
	material.UpdatedAt = time.Now()
	if err := s.materialRepo.UpdateWithTx(ctx, tx, material); err != nil {
		return errors.NewInternalError("failed to rollback material status", err)
	}

	// Update operation status to failed
	operation.Status = entities.OperationStatusFailed
	operation.Timestamp = time.Now()
	if err := s.operationRepo.UpdateWithTx(ctx, tx, operation); err != nil {
		return errors.NewInternalError("failed to update operation status to failed", err)
	}

	// Publish physical placement failed event
	s.publishPhysicalPlacementFailedEvent(ctx, operation)

	if err := tx.Commit(); err != nil {
		return errors.NewInternalError("failed to commit transaction", err.Error)
	}

	return nil
}

func (s *InventoryService) HandlePhysicalRemovalTimeout(ctx context.Context, operationID string) error {
	operation, err := s.operationRepo.GetByID(ctx, operationID)
	if err != nil {
		return errors.NewNotFoundError("operation not found", err)
	}

	if operation.Status != entities.OperationStatusPendingRemovalConfirmation {
		return errors.NewConflictError(fmt.Sprintf("operation %s is not in pending physical confirmation status", operationID), nil)
	}

	tx, err := s.operationRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	slot, err := s.slotRepo.GetByID(ctx, operation.SlotID)
	if err != nil {
		return errors.NewNotFoundError("slot not found for rollback", err)
	}
	slot.Status = entities.SlotStatusOccupied // Rollback to occupied status
	slot.UpdatedAt = time.Now()
	slot.Version++
	if err := s.slotRepo.UpdateWithTx(ctx, tx, slot); err != nil {
		return errors.NewInternalError("failed to rollback slot status", err)
	}

	// Update operation status to failed
	operation.Status = entities.OperationStatusFailed
	operation.Timestamp = time.Now()
	if err := s.operationRepo.UpdateWithTx(ctx, tx, operation); err != nil {
		return errors.NewInternalError("failed to update operation status to failed", err)
	}

	// Publish physical removal failed event
	s.publishPhysicalRemovalFailedEvent(ctx, operation)

	return nil
}

// HandleMaterialDetectedEvent handles a material detected event from a physical sensor.
// It checks if this detection confirms a pending placement operation or is an unplanned placement.
func (s *InventoryService) HandleMaterialDetectedEvent(ctx context.Context, slotID, materialBarcode string) error {
	// First, try to find a pending physical confirmation operation for this slot
	operations, err := s.operationRepo.GetPendingPhysicalConfirmationsBySlotID(ctx, slotID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to query pending operations for slot %s: %v", slotID, err), err)
		return err
	}

	for _, op := range operations {
		// Check if the detected material matches the expected material for this operation
		if op.MaterialID == materialBarcode { // Assuming materialBarcode from sensor matches MaterialID in operation
			logger.Info(fmt.Sprintf("Confirming physical placement for operation %s in slot %s", op.ID, slotID))
			return s.ConfirmPhysicalPlacement(ctx, op.ID)
		}
	}

	// If no matching pending operation is found, it's an unplanned placement
	logger.Info(fmt.Sprintf("Unplanned material detected in slot %s with barcode %s. Triggering alert.", slotID, materialBarcode))
	s.publishUnplannedPlacementEvent(ctx, slotID, materialBarcode)

	return nil
}

func (s *InventoryService) HandleMaterialRemovedEvent(ctx context.Context, slotID string, materialBarcode string) error {
	operations, err := s.operationRepo.GetPendingRemovalConfirmationsBySlotID(ctx, slotID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to query pending operations for slot %s: %v", slotID, err), err)
		return err
	}

	for _, op := range operations {
		if op.Status == entities.OperationStatusPendingRemovalConfirmation {
			logger.Info(fmt.Sprintf("Confirming removal for operation %s in slot %s", op.ID, slotID))
			return s.ConfirmPhysicalRemoval(ctx, op.ID, false)
		}
	}

	logger.Info(fmt.Sprintf("Unplanned removal detected in slot %s with barcode %s. Triggering alert.", slotID, materialBarcode))
	s.publishUnplannedRemovalEvent(ctx, slotID, materialBarcode)

	return s.ConfirmPhysicalRemoval(ctx, slotID, true)
}