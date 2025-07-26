package services

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "warehouse/internal/domain/entities"
    "warehouse/internal/domain/repositories"
    "warehouse/pkg/errors"
)

type InventoryService struct {
    materialRepo   repositories.MaterialRepository
    slotRepo       repositories.SlotRepository
    operationRepo  repositories.OperationRepository
    lockService    *LockService
    eventService   *EventService
}

func NewInventoryService(
    materialRepo repositories.MaterialRepository,
    slotRepo repositories.SlotRepository,
    operationRepo repositories.OperationRepository,
    lockService *LockService,
    eventService *EventService,
) *InventoryService {
    return &InventoryService{
        materialRepo:  materialRepo,
        slotRepo:      slotRepo,
        operationRepo: operationRepo,
        lockService:   lockService,
        eventService:  eventService,
    }
}

type PlaceMaterialCommand struct {
    MaterialBarcode string `json:"material_barcode" validate:"required"`
    SlotID         string `json:"slot_id" validate:"required"`
    OperatorID     string `json:"operator_id" validate:"required"`
}

func (s *InventoryService) PlaceMaterial(ctx context.Context, cmd PlaceMaterialCommand) error {
    // 1. 獲取分佈式鎖
    lockKey := fmt.Sprintf("slot:%s", cmd.SlotID)
    unlock, err := s.lockService.AcquireLock(ctx, lockKey, 30*time.Second)
    if err != nil {
        return errors.NewConflictError("slot is being modified", err)
    }
    defer unlock()
    
    // 2. 驗證格子狀態
    slot, err := s.slotRepo.GetByID(ctx, cmd.SlotID)
    if err != nil {
        return errors.NewNotFoundError("slot not found", err)
    }
    
    if slot.Status != entities.SlotStatusEmpty {
        return errors.NewConflictError("slot is not available", nil)
    }
    
    // 3. 驗證材料狀態
    material, err := s.materialRepo.GetByBarcode(ctx, cmd.MaterialBarcode)
    if err != nil {
        return errors.NewNotFoundError("material not found", err)
    }
    
    if material.Status != entities.MaterialStatusAvailable {
        return errors.NewConflictError("material is not available", nil)
    }
    
    // 4. 開始事務
    tx, err := s.slotRepo.BeginTx(ctx)
    if err != nil {
        return errors.NewInternalError("failed to start transaction", err)
    }
    defer tx.Rollback()
    
    // 5. 更新格子狀態
    slot.Status = entities.SlotStatusOccupied
    slot.MaterialID = &material.ID
    slot.UpdatedAt = time.Now()
    slot.Version++
    
    if err := s.slotRepo.UpdateWithTx(ctx, tx, slot); err != nil {
        return errors.NewConflictError("failed to update slot", err)
    }
    
    // 6. 更新材料狀態
    material.Status = entities.MaterialStatusInUse
    material.UpdatedAt = time.Now()
    
    if err := s.materialRepo.UpdateWithTx(ctx, tx, material); err != nil {
        return errors.NewInternalError("failed to update material", err)
    }
    
    // 7. 記錄操作
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
        return errors.NewInternalError("failed to record operation", err)
    }
    
    // 8. 提交事務
    if err := tx.Commit(); err != nil {
        return errors.NewInternalError("failed to commit transaction", err)
    }
    
    // 9. 發布事件
    event := &events.MaterialPlacedEvent{
        EventID:    generateUUID(),
        MaterialID: material.ID,
        SlotID:     cmd.SlotID,
        ShelfID:    slot.ShelfID,
        OperatorID: cmd.OperatorID,
        Timestamp:  time.Now(),
    }
    
    if err := s.eventService.PublishEvent(ctx, "material.placed", event); err != nil {
        // 事件發布失敗不影響主流程，但需要記錄日誌
        logger.Error("Failed to publish event", "error", err, "event", event)
    }
    
    return nil
}

func (s *InventoryService) GetShelfStatus(ctx context.Context, shelfID string) (*entities.ShelfStatus, error) {
    // 先嘗試從緩存獲取
    if status := s.getShelfStatusFromCache(ctx, shelfID); status != nil {
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
        Slots:         slots,
        UpdatedAt:     time.Now(),
    }
    
    for _, slot := range slots {
        switch slot.Status {
        case entities.SlotStatusEmpty:
            status.EmptySlots++
        case entities.SlotStatusOccupied:
            status.OccupiedSlots++
        }
    }
    
    // 更新緩存
    s.cacheShelfStatus(ctx, shelfID, status)
    
    return status, nil
}