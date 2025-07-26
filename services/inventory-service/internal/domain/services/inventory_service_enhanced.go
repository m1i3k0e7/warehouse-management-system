package services

import (
    "context"
    "fmt"
    "time"
    
    "warehouse/internal/domain/entities"
    "warehouse/internal/domain/repositories"
    "warehouse/pkg/errors"
    "warehouse/pkg/logger"
)

type InventoryService struct {
    materialRepo       repositories.MaterialRepository
    slotRepo          repositories.SlotRepository
    operationRepo     repositories.OperationRepository
    alertRepo         repositories.AlertRepository
    lockService       *LockService
    eventService      *EventService
    cacheService      *CacheService
    auditService      *AuditService
    alertService      *AlertService
}

type SensorData struct {
    Weight      float64 `json:"weight"`
    Temperature float64 `json:"temperature"`
    Humidity    float64 `json:"humidity"`
}

type PlaceMaterialCommand struct {
    MaterialBarcode string      `json:"material_barcode" validate:"required"`
    SlotID         string      `json:"slot_id" validate:"required"`
    OperatorID     string      `json:"operator_id" validate:"required"`
    SensorData     *SensorData `json:"sensor_data,omitempty"`
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
) *InventoryService {
    return &InventoryService{
        materialRepo:  materialRepo,
        slotRepo:     slotRepo,
        operationRepo: operationRepo,
        alertRepo:    alertRepo,
        lockService:  lockService,
        eventService: eventService,
        cacheService: cacheService,
        auditService: auditService,
        alertService: alertService,
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
        ShelfID:           shelfID,
        TotalSlots:        len(slots),
        HealthySlots:      0,
        ErrorSlots:        0,
        MaintenanceSlots:  0,
        LastCheckTime:     time.Now(),
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
