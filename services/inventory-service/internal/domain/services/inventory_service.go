
package services

import (
	"context"
	"fmt"
	"time"
    "encoding/json"

	"warehouse/internal/domain/entities"
	"warehouse/internal/domain/repositories"
	"warehouse/pkg/errors"
	"warehouse/pkg/logger"
)

type InventoryService struct {
	materialRepo  repositories.MaterialRepository
	slotRepo      repositories.SlotRepository
	operationRepo repositories.OperationRepository
	alertRepo     repositories.AlertRepository
	lockService   *LockService
	eventService  *EventService
	cacheService  *CacheService
	auditService  *AuditService
	alertService  *AlertService
	failedEventRepo repositories.FailedEventRepository
}

type SensorData struct {
	Weight      float64 `json:"weight"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

type PlaceMaterialCommand struct {
	MaterialBarcode string      `json:"material_barcode" validate:"required"`
	SlotID          string      `json:"slot_id" validate:"required"`
	OperatorID      string      `json:"operator_id" validate:"required"`
	SensorData      *SensorData `json:"sensor_data,omitempty"`
}

type RemoveMaterialCommand struct {
	SlotID     string `json:"slot_id" validate:"required"`
	OperatorID string `json:"operator_id" validate:"required"`
	Reason     string `json:"reason,omitempty"`
}

type MoveMaterialCommand struct {
	FromSlotID string `json:"from_slot_id" validate:"required"`
	ToSlotID   string `json:"to_slot_id" validate:"required"`
	OperatorID string `json:"operator_id" validate:"required"`
	Reason     string `json:"reason,omitempty"`
}

type ReserveSlotsCommand struct {
	SlotIDs    []string `json:"slot_ids" validate:"required"`
	OperatorID string   `json:"operator_id" validate:"required"`
	Duration   int      `json:"duration"` // 預留時間（分鐘）
	Purpose    string   `json:"purpose"`
}

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
		materialRepo:  materialRepo,
		slotRepo:      slotRepo,
		operationRepo: operationRepo,
		alertRepo:     alertRepo,
		lockService:   lockService,
		eventService:  eventService,
		cacheService:  cacheService,
		auditService:  auditService,
		alertService:  alertService,
		failedEventRepo: failedEventRepo,
	}
}

// 增強版材料放置功能
func (s *InventoryService) PlaceMaterial(ctx context.Context, cmd PlaceMaterialCommand) error {
	// 1. 參數驗證
	if err := s.validatePlaceMaterialCommand(cmd); err != nil {
		return errors.NewValidationError("invalid command", err)
	}

	// 2. 獲取料架級別鎖
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

	// 3. 業務邏輯驗證
	if err := s.validatePlacementPreconditions(ctx, cmd); err != nil {
		return err
	}

	// 4. 執行放置操作
	operation, err := s.executePlaceMaterial(ctx, cmd)
	if err != nil {
		// 記錄失敗操作
		s.auditService.LogFailedOperation(ctx, "place_material", cmd, err)
		return err
	}

	// 5. 檢查異常情況
	s.checkForAnomalies(ctx, operation, cmd.SensorData)

	// 6. 記錄審計日誌
	s.auditService.LogSuccessfulOperation(ctx, operation)

	return nil
}

// 材料移除功能
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

// 材料移動功能
func (s *InventoryService) MoveMaterial(ctx context.Context, cmd MoveMaterialCommand) error {
	// 獲取源格子和目標格子
	fromSlot, err := s.slotRepo.GetByID(ctx, cmd.FromSlotID)
	if err != nil {
		return errors.NewNotFoundError("source slot not found", err)
	}

	toSlot, err := s.slotRepo.GetByID(ctx, cmd.ToSlotID)
	if err != nil {
		return errors.NewNotFoundError("target slot not found", err)
	}

	// 驗證移動條件
	if fromSlot.Status != entities.SlotStatusOccupied {
		return errors.NewConflictError("source slot is empty", nil)
	}

	if toSlot.Status != entities.SlotStatusEmpty {
		return errors.NewConflictError("target slot is not empty", nil)
	}

	// 如果跨料架移動，需要獲取兩個鎖
	locks := s.acquireMultipleShelfLocks(ctx, []string{fromSlot.ShelfID, toSlot.ShelfID})
	defer s.releaseMultipleLocks(locks)

	return s.executeMoveMaterial(ctx, cmd, fromSlot, toSlot)
}

// 格子預留功能
func (s *InventoryService) ReserveSlots(ctx context.Context, cmd ReserveSlotsCommand) error {
	// 獲取所有相關料架的鎖
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

// 智能尋找最佳格子
func (s *InventoryService) FindOptimalSlot(ctx context.Context, materialType string, shelfID string) (*entities.Slot, error) {
	slots, err := s.slotRepo.GetEmptySlotsByShelf(ctx, shelfID)
	if err != nil {
		return nil, err
	}

	if len(slots) == 0 {
		return nil, errors.NewNotFoundError("no empty slots available", nil)
	}

	// 智能選擇邏輯：
	// 1. 優先選擇同類型材料附近的格子
	// 2. 考慮人機工程學（中間位置優先）
	// 3. 避免過於集中

	return s.selectBestSlot(slots, materialType)
}

// 批量操作功能
func (s *InventoryService) BatchPlaceMaterials(ctx context.Context, commands []PlaceMaterialCommand) error {
	// 按料架分組
	shelfGroups := s.groupCommandsByShelf(commands)

	for shelfID, shelfCommands := range shelfGroups {
		lockKey := fmt.Sprintf("shelf:%s", shelfID)
		unlock, err := s.lockService.AcquireLock(ctx, lockKey, 60*time.Second)
		if err != nil {
			return errors.NewConflictError(fmt.Sprintf("failed to lock shelf %s", shelfID), err)
		}

		err = s.executeBatchPlacement(ctx, shelfCommands)
		unlock()

		if err != nil {
			return err
		}
	}

	return nil
}

// 料架健康檢查
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

	// 如果健康分數低於閾值，發送告警
	if health.HealthScore < 95.0 {
		s.alertService.SendShelfHealthAlert(ctx, health)
	}

	return health, nil
}

// 處理格子錯誤
func (s *InventoryService) HandleSlotError(ctx context.Context, slotID string, errorType string) error {
	slot, err := s.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		return err
	}

	// 記錄錯誤
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

	// 根據錯誤類型決定處理方式
	switch errorType {
	case "sensor_error":
		return s.markSlotForMaintenance(ctx, slotID, "sensor malfunction")
	case "weight_mismatch":
		return s.triggerManualVerification(ctx, slotID)
	default:
		return s.markSlotForInvestigation(ctx, slotID, errorType)
	}
}

// 更新料架狀態
func (s *InventoryService) UpdateShelfStatus(ctx context.Context, shelfID string, status string) error {
	// 更新緩存中的料架狀態
	cacheKey := fmt.Sprintf("shelf_status:%s", shelfID)
	statusData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	return s.cacheService.Set(ctx, cacheKey, statusData, time.Hour)
}

func (s *InventoryService) GetShelfStatus(ctx context.Context, shelfID string) (*entities.ShelfStatus, error) {
	// 先嘗試從緩存獲取
	if status, err := s.cacheService.GetShelfStatus(ctx, shelfID); err == nil && status != nil {
		return status, nil
	}

	// 從數據庫查詢
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

	// 更新緩存
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
	defer tx.Rollback()

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
		Status:     entities.OperationStatusCompleted,
	}

	if err := s.operationRepo.CreateWithTx(ctx, tx, operation); err != nil {
		return nil, errors.NewInternalError("failed to record operation", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.NewInternalError("failed to commit transaction", err)
	}

	s.publishMaterialPlacedEvent(ctx, operation)

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
	defer tx.Rollback()

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
	defer tx.Rollback()

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
	defer tx.Rollback()

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
	if len(slots) > 0 {
		return slots[0], nil
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
	defer tx.Rollback()

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

func generateUUID() string {
    // In a real application, you would use a library like github.com/google/uuid
    return "some-uuid"
}
