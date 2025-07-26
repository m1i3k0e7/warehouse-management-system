package database

import (
	"fmt"
	"inventory-service/internal/domain/entities"
	"inventory-service/pkg/utils"
	"time"
	
	"gorm.io/gorm"
)

func SeedData(db *gorm.DB) error {
	// 檢查是否已經有數據
	var count int64
	db.Model(&entities.Material{}).Count(&count)
	if count > 0 {
		return nil // 已有數據，跳過初始化
	}
	
	// 創建模擬材料數據
	materials := generateMockMaterials(500)
	if err := db.CreateInBatches(materials, 100).Error; err != nil {
		return fmt.Errorf("failed to seed materials: %w", err)
	}
	
	// 創建模擬料架和格子數據
	slots := generateMockSlots(100, 700) // 100個料架，每個700個格子
	if err := db.CreateInBatches(slots, 1000).Error; err != nil {
		return fmt.Errorf("failed to seed slots: %w", err)
	}
	
	// 隨機分配一些材料到格子中
	if err := assignRandomMaterials(db, materials, slots); err != nil {
		return fmt.Errorf("failed to assign materials: %w", err)
	}
	
	return nil
}

func generateMockMaterials(count int) []*entities.Material {
	materials := make([]*entities.Material, count)
	materialTypes := []string{"IC", "Resistor", "Capacitor", "Inductor", "Connector", "CPU", "Memory", "PCB"}
	
	for i := 0; i < count; i++ {
		materials[i] = &entities.Material{
			ID:        utils.GenerateUUID(),
			Barcode:   fmt.Sprintf("MAT%06d", i+1),
			Name:      fmt.Sprintf("%s Component %d", materialTypes[i%len(materialTypes)], i+1),
			Type:      materialTypes[i%len(materialTypes)],
			Status:    entities.MaterialStatusAvailable,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
	
	return materials
}

func generateMockSlots(shelfCount, slotsPerShelf int) []*entities.Slot {
	totalSlots := shelfCount * slotsPerShelf
	slots := make([]*entities.Slot, totalSlots)
	
	index := 0
	for shelfNum := 1; shelfNum <= shelfCount; shelfNum++ {
		rows := 25    // 每個料架25行
		cols := 28    // 每個料架28列
		
		for row := 1; row <= rows; row++ {
			for col := 1; col <= cols; col++ {
				if index >= totalSlots {
					break
				}
				
				slots[index] = &entities.Slot{
					ID:        fmt.Sprintf("SHELF%03d-R%02d-C%02d", shelfNum, row, col),
					ShelfID:   fmt.Sprintf("SHELF%03d", shelfNum),
					Row:       row,
					Column:    col,
					Status:    entities.SlotStatusEmpty,
					UpdatedAt: time.Now(),
					Version:   1,
				}
				index++
			}
		}
	}
	
	return slots[:index]
}

func assignRandomMaterials(db *gorm.DB, materials []*entities.Material, slots []*entities.Slot) error {
	// 隨機分配約30%的材料到格子中
	assignCount := len(materials) * 3 / 10
	
	for i := 0; i < assignCount && i < len(slots); i++ {
		material := materials[i]
		slot := slots[i*2] // 間隔分配，避免太密集
		
		// 更新格子狀態
		slot.Status = entities.SlotStatusOccupied
		slot.MaterialID = &material.ID
		
		// 更新材料狀態
		material.Status = entities.MaterialStatusInUse
		
		// 保存到數據庫
		if err := db.Save(slot).Error; err != nil {
			return err
		}
		if err := db.Save(material).Error; err != nil {
			return err
		}
		
		// 創建操作記錄
		operation := &entities.Operation{
			ID:         utils.GenerateUUID(),
			Type:       entities.OperationTypePlacement,
			MaterialID: material.ID,
			SlotID:     slot.ID,
			OperatorID: "SYSTEM",
			ShelfID:    slot.ShelfID,
			Timestamp:  time.Now(),
			Status:     entities.OperationStatusCompleted,
		}
		
		if err := db.Create(operation).Error; err != nil {
			return err
		}
	}
	
	return nil
}