package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"warehouse/internal/domain/entities"
	"warehouse/internal/domain/repositories"
	"warehouse/pkg/errors"
	"warehouse/pkg/logger"
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

func (s *InventoryService) PlaceMaterial(ctx context.Context, cmd PlaceMaterialCommand) error {
	// validate command parameters
	if err := s.validatePlaceMaterialCommand(cmd); err != nil {
		return errors.NewValidationError("invalid command", err)
	}

	// acquire lock on the shelf
	slot, err := s.slotRepo.GetByID(ctx, cmd.SlotID)
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
	if err := s.validatePlacementPreconditions(ctx, cmd); err != nil {
		return err
	}

	// execute the placement operation
	operation, err := s.executePlaceMaterial(ctx, cmd)
	if err != nil {
		// log the failed operation
		s.auditService.LogFailedOperation(ctx, "place_material", cmd, err)
		return err
	}

	// check for anomalies
	s.checkForAnomalies(ctx, operation, cmd.SensorData)

	// log the successful operation
	s.auditService.LogSuccessfulOperation(ctx, operation)
	
	return nil
}

func (s *InventoryService) RemoveMaterial(ctx context.Context, cmd RemoveMaterialCommand) error {
	slot, err := s.slotRepo.GetByID(ctx, cmd.SlotID)
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

	return s.executeRemoveMaterial(ctx, cmd, slot)
}

func (s *InventoryService) MoveMaterial(ctx context.Context, cmd MoveMaterialCommand) error {
	// get the source and target slots
	fromSlot, err := s.slotRepo.GetByID(ctx, cmd.FromSlotID)
	if err != nil {
		return errors.NewNotFoundError("source slot not found", err)
	}

	toSlot, err := s.slotRepo.GetByID(ctx, cmd.ToSlotID)
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

	return s.executeMoveMaterial(ctx, cmd, fromSlot, toSlot)
}

func (s *InventoryService) ReserveSlots(ctx context.Context, cmd ReserveSlotsCommand) error {
	// acquire locks on all shelves involved in the reservation
	shelfIDs := make([]string, 0)
	slotShelfMap := make(map[string]string)
	for _, slotID := range cmd.SlotIDs {
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

	return s.executeReserveSlots(ctx, cmd, slotShelfMap)
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

func (s *InventoryService) BatchPlaceMaterials(ctx context.Context, commands []PlaceMaterialCommand) error {
	// group commands by shelf
	shelfGroups := s.groupCommandsByShelf(commands)
	for shelfID, shelfCommands := range shelfGroups {
		lockKey := fmt.Sprintf("shelf:%s", shelfID)
		unlock, err := s.lockService.AcquireLock(ctx, lockKey, 60*time.Second)
		if err != nil {
			return errors.NewConflictError(fmt.Sprintf("failed to lock shelf %s", shelfID), err)
		}
		defer unlock()

		err = s.executeBatchPlacement(ctx, shelfCommands)
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

func (s *InventoryService) validatePlaceMaterialCommand(cmd PlaceMaterialCommand) error {
	if cmd.MaterialBarcode == "" || cmd.SlotID == "" || cmd.OperatorID == "" {
		return fmt.Errorf("material barcode, slot ID, and operator ID are required")
	}

	return nil
}

func (s *InventoryService) validatePlacementPreconditions(ctx context.Context, cmd PlaceMaterialCommand) error {
	slot, err := s.slotRepo.GetByID(ctx, cmd.SlotID)
	if err != nil {
		return errors.NewNotFoundError("slot not found", err)
	}
	if slot.Status != entities.SlotStatusEmpty {
		return errors.NewConflictError("slot is not available", nil)
	}

	material, err := s.materialRepo.GetByBarcode(ctx, cmd.MaterialBarcode)
	if err != nil {
		return errors.NewNotFoundError("material not found", err)
	}
	if material.Status != entities.MaterialStatusAvailable {
		return errors.NewConflictError("material is not available", nil)
	}

	return nil
}

func (s *InventoryService) executePlaceMaterial(ctx context.Context, cmd PlaceMaterialCommand) (*entities.Operation, error) {
	tx, err := s.slotRepo.BeginTx(ctx)
	if err != nil {
		return nil, errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	slot, _ := s.slotRepo.GetByID(ctx, cmd.SlotID)
	material, _ := s.materialRepo.GetByBarcode(ctx, cmd.MaterialBarcode)
	slot.Status = entities.SlotStatusOccupied
	slot.MaterialID = &material.ID
	slot.UpdatedAt = time.Now()
	slot.Version++
	if err := s.slotRepo.UpdateWithTx(ctx, tx, slot); err != nil {
		return nil, errors.NewConflictError("failed to update slot", err)
	}

	material.Status = entities.MaterialStatusInUse
	material.UpdatedAt = time.Now()
	if err := s.materialRepo.UpdateWithTx(ctx, tx, material); err != nil {
		return nil, errors.NewInternalError("failed to update material", err)
	}

	operation := &entities.Operation{
		ID:         generateUUID(),
		Type:       entities.OperationTypePlacement,
		MaterialID: material.ID,
		SlotID:     cmd.SlotID,
		OperatorID: cmd.OperatorID,
		ShelfID:    slot.ShelfID,
		Timestamp:  time.Now(),
		Status:     entities.OperationStatusPendingPhysicalConfirmation,
	}
	if err := s.operationRepo.CreateWithTx(ctx, tx, operation); err != nil {
		return nil, errors.NewInternalError("failed to record operation", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.NewInternalError("failed to commit transaction", err)
	}

	// Publish event to request physical placement
	s.publishPhysicalPlacementRequestedEvent(ctx, operation)
	return operation, nil
}

func (s *InventoryService) checkForAnomalies(ctx context.Context, operation *entities.Operation, sensorData *SensorData) {
	if sensorData != nil {
		// Example anomaly detection: check if the weight is within a reasonable range
		// This is a simplified example. A real implementation would be more complex.
		if sensorData.Weight > 1000 { // Assuming weight is in grams
			s.publishSystemAlertEvent(ctx, "weight_anomaly", "high", "Anomalous weight detected", map[string]interface{}{
				"operation_id": operation.ID,
				"weight":       sensorData.Weight,
			})
		}
	}
}

func (s *InventoryService) executeRemoveMaterial(ctx context.Context, cmd RemoveMaterialCommand, slot *entities.Slot) error {
	tx, err := s.slotRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	material, err := s.materialRepo.GetByID(ctx, *slot.MaterialID)
	if err != nil {
		return errors.NewNotFoundError("material not found", err)
	}

	slot.Status = entities.SlotStatusEmpty
	slot.MaterialID = nil
	slot.UpdatedAt = time.Now()
	slot.Version++
	if err := s.slotRepo.UpdateWithTx(ctx, tx, slot); err != nil {
		return errors.NewConflictError("failed to update slot", err)
	}

	material.Status = entities.MaterialStatusAvailable
	material.UpdatedAt = time.Now()
	if err := s.materialRepo.UpdateWithTx(ctx, tx, material); err != nil {
		return errors.NewInternalError("failed to update material", err)
	}

	operation := &entities.Operation{
		ID:         generateUUID(),
		Type:       entities.OperationTypeRemoval,
		MaterialID: material.ID,
		SlotID:     cmd.SlotID,
		OperatorID: cmd.OperatorID,
		ShelfID:    slot.ShelfID,
		Timestamp:  time.Now(),
		Status:     entities.OperationStatusCompleted,
	}
	if err := s.operationRepo.CreateWithTx(ctx, tx, operation); err != nil {
		return errors.NewInternalError("failed to record operation", err)
	}

	if err := tx.Commit(); err != nil {
		return errors.NewInternalError("failed to commit transaction", err)
	}

	s.publishMaterialRemovedEvent(ctx, operation)

	return nil
}

func (s *InventoryService) executeMoveMaterial(ctx context.Context, cmd MoveMaterialCommand, fromSlot, toSlot *entities.Slot) error {
	tx, err := s.slotRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

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
		SlotID:     cmd.ToSlotID,
		OperatorID: cmd.OperatorID,
		ShelfID:    toSlot.ShelfID,
		Timestamp:  time.Now(),
		Status:     entities.OperationStatusCompleted,
	}
	if err := s.operationRepo.CreateWithTx(ctx, tx, operation); err != nil {
		return errors.NewInternalError("failed to record operation", err)
	}

	if err := tx.Commit(); err != nil {
		return errors.NewInternalError("failed to commit transaction", err)
	}

	s.publishMaterialMovedEvent(ctx, operation, cmd.FromSlotID)

	return nil
}

func (s *InventoryService) executeReserveSlots(ctx context.Context, cmd ReserveSlotsCommand, slotShelfMap map[string]string) error {
	tx, err := s.slotRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	for _, slotID := range cmd.SlotIDs {
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
		return errors.NewInternalError("failed to commit transaction", err)
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

func (s *InventoryService) groupCommandsByShelf(commands []PlaceMaterialCommand) (map[string][]PlaceMaterialCommand, error) {
	shelfGroups := make(map[string][]PlaceMaterialCommand)
	for _, cmd := range commands {
		slot, err := s.slotRepo.GetByID(context.Background(), cmd.SlotID)
		if err != nil {
			return nil, errors.NewNotFoundError(fmt.Sprintf("slot %s not found", cmd.SlotID), err)
		}
		shelfGroups[slot.ShelfID] = append(shelfGroups[slot.ShelfID], cmd)
	}
	return shelfGroups, nil
}

func (s *InventoryService) executeBatchPlacement(ctx context.Context, commands []PlaceMaterialCommand) error {
	tx, err := s.slotRepo.BeginTx(ctx)
	if err != nil {
		return errors.NewInternalError("failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	for _, cmd := range commands {
		_, err := s.executePlaceMaterial(ctx, cmd)
		if err != nil {
			return err // Or collect errors and return them all
		}
	}
	return tx.Commit()
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

func (s *InventoryService) searchMaterials(ctx context.Context, query string, limit, offset int) ([]*entities.Material, error) {
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
		return errors.NewInternalError("failed to commit transaction", err)
	}

	// Publish material placed event (now that physical placement is confirmed)
	s.publishMaterialPlacedEvent(ctx, operation)
	// s.publishPhysicalPlacementConfirmedEvent(ctx, operation) // This event is now handled by material.placed

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
		return errors.NewInternalError("failed to commit transaction", err)
	}

	return nil
}

// HandleMaterialDetectedEvent handles a material detected event from a physical sensor.
// It checks if this detection confirms a pending placement operation or is an unplanned placement.
func (s *InventoryService) HandleMaterialDetectedEvent(ctx context.Context, slotID, materialBarcode string) error {
	// First, try to find a pending physical confirmation operation for this slot
	operations, err := s.operationRepo.GetPendingPhysicalConfirmationsBySlotID(ctx, slotID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to query pending operations for slot %s: %v", slotID, err))
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
	logger.Warn(fmt.Sprintf("Unplanned material detected in slot %s with barcode %s. Triggering alert.", slotID, materialBarcode))
	s.publishUnplannedPlacementEvent(ctx, slotID, materialBarcode)

	return nil
}

func (s *InventoryService) publishPhysicalPlacementRequestedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		OperationID string    `json:"operation_id"`
		MaterialID  string    `json:"material_id"`
		SlotID      string    `json:"slot_id"`
		ShelfID     string    `json:"shelf_id"`
		OperatorID  string    `json:"operator_id"`
		Timestamp   time.Time `json:"timestamp"`
	}{
		OperationID: operation.ID,
		MaterialID:  operation.MaterialID,
		SlotID:      operation.SlotID,
		ShelfID:     operation.ShelfID,
		OperatorID:  operation.OperatorID,
		Timestamp:   time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypePhysicalPlacementRequested, event); err != nil {
		logger.Error("Failed to publish physical placement requested event", err)
		s.scheduleEventRetry(ctx, EventTypePhysicalPlacementRequested, EventTypePhysicalPlacementRequested, event, err)
	}
}

func (s *InventoryService) publishPhysicalPlacementConfirmedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		OperationID string    `json:"operation_id"`
		MaterialID  string    `json:"material_id"`
		SlotID      string    `json:"slot_id"`
		ShelfID     string    `json:"shelf_id"`
		OperatorID  string    `json:"operator_id"`
		Timestamp   time.Time `json:"timestamp"`
	}{
		OperationID: operation.ID,
		MaterialID:  operation.MaterialID,
		SlotID:      operation.SlotID,
		ShelfID:     operation.ShelfID,
		OperatorID:  operation.OperatorID,
		Timestamp:   time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypePhysicalPlacementConfirmed, event); err != nil {
		logger.Error("Failed to publish physical placement confirmed event", err)
		s.scheduleEventRetry(ctx, EventTypePhysicalPlacementConfirmed, EventTypePhysicalPlacementConfirmed, event, err)
	}
}

func (s *InventoryService) publishPhysicalPlacementFailedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		OperationID string    `json:"operation_id"`
		MaterialID  string    `json:"material_id"`
		SlotID      string    `json:"slot_id"`
		ShelfID     string    `json:"shelf_id"`
		OperatorID  string    `json:"operator_id"`
		Timestamp   time.Time `json:"timestamp"`
	}{
		OperationID: operation.ID,
		MaterialID:  operation.MaterialID,
		SlotID:      operation.SlotID,
		ShelfID:     operation.ShelfID,
		OperatorID:  operation.OperatorID,
		Timestamp:   time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypePhysicalPlacementFailed, event); err != nil {
		logger.Error("Failed to publish physical placement failed event", err)
		s.scheduleEventRetry(ctx, EventTypePhysicalPlacementFailed, EventTypePhysicalPlacementFailed, event, err)
	}
}

func (s *InventoryService) publishUnplannedPlacementEvent(ctx context.Context, slotID, materialBarcode string) {
	event := struct {
		SlotID        string    `json:"slot_id"`
		MaterialBarcode string    `json:"material_barcode"`
		Timestamp     time.Time `json:"timestamp"`
	}{
		SlotID:        slotID,
		MaterialBarcode: materialBarcode,
		Timestamp:     time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeUnplannedPlacement, event); err != nil {
		logger.Error("Failed to publish unplanned placement event", err)
		s.scheduleEventRetry(ctx, EventTypeUnplannedPlacement, EventTypeUnplannedPlacement, event, err)
	}
}

func (s *InventoryService) publishMaterialPlacedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		EventID    string    `json:"event_id"`
		MaterialID string    `json:"material_id"`
		SlotID     string    `json:"slot_id"`
		ShelfID    string    `json:"shelf_id"`
		OperatorID string    `json:"operator_id"`
		Timestamp  time.Time `json:"timestamp"`
	}{
		EventID:    generateUUID(),
		MaterialID: operation.MaterialID,
		SlotID:     operation.SlotID,
		ShelfID:    operation.ShelfID,
		OperatorID: operation.OperatorID,
		Timestamp:  time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeMaterialPlaced, event); err != nil {
		logger.Error("Failed to publish material placed event", err)
		s.scheduleEventRetry(ctx, EventTypeMaterialPlaced, EventTypeMaterialPlaced, event, err)
	}
}

func (s *InventoryService) publishMaterialRemovedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		EventID    string    `json:"event_id"`
		MaterialID string    `json:"material_id"`
		SlotID     string    `json:"slot_id"`
		ShelfID    string    `json:"shelf_id"`
		OperatorID string    `json:"operator_id"`
		Timestamp  time.Time `json:"timestamp"`
	}{
		EventID:    generateUUID(),
		MaterialID: operation.MaterialID,
		SlotID:     operation.SlotID,
		ShelfID:    operation.ShelfID,
		OperatorID: operation.OperatorID,
		Timestamp:  time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeMaterialRemoved, event); err != nil {
		logger.Error("Failed to publish material removed event", err)
		s.scheduleEventRetry(ctx, EventTypeMaterialRemoved, EventTypeMaterialRemoved, event, err)
	}
}

func (s *InventoryService) publishMaterialMovedEvent(ctx context.Context, operation *entities.Operation, fromSlotID string) {
	event := struct {
		EventID    string    `json:"event_id"`
		MaterialID string    `json:"material_id"`
		FromSlotID string    `json:"from_slot_id"`
		ToSlotID   string    `json:"to_slot_id"`
		ShelfID    string    `json:"shelf_id"`
		OperatorID string    `json:"operator_id"`
		Timestamp  time.Time `json:"timestamp"`
	}{
		EventID:    generateUUID(),
		MaterialID: operation.MaterialID,
		FromSlotID: fromSlotID,
		ToSlotID:   operation.SlotID,
		ShelfID:    operation.ShelfID,
		OperatorID: operation.OperatorID,
		Timestamp:  time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeMaterialMoved, event); err != nil {
		logger.Error("Failed to publish material moved event", err)
		s.scheduleEventRetry(ctx, EventTypeMaterialMoved, EventTypeMaterialMoved, event, err)
	}
}
