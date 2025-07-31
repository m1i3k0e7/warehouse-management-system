package database

import (
	"fmt"
	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/pkg/utils"
	"time"
	
	"gorm.io/gorm"
)

func SeedData(db *gorm.DB) error {
	// check if the database already has data
	var count int64
	db.Model(&entities.Material{}).Count(&count)
	if count > 0 {
		return nil // skip seeding if data already exists
	}
	
	// generating mock materials
	materials := generateMockMaterials(500)
	if err := db.CreateInBatches(materials, 100).Error; err != nil {
		return fmt.Errorf("failed to seed materials: %w", err)
	}
	
	// generating mock slots
	slots := generateMockSlots(100, 700) // 100 shelves, 700 slots per shelf
	if err := db.CreateInBatches(slots, 1000).Error; err != nil {
		return fmt.Errorf("failed to seed slots: %w", err)
	}
	
	// randomly assign materials to slots
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
		rows := 7    // 7 rows per shelf
		cols := 100  // 100 columns per shelf
		
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
	// randomly assign 30% of materials to slots
	assignCount := len(materials) * 3 / 10
	
	for i := 0; i < assignCount && i < len(slots); i++ {
		material := materials[i]
		slot := slots[i*2] // assign every second slot to avoid overloading
		
		// update slot with material
		slot.Status = entities.SlotStatusOccupied
		slot.MaterialID = &material.ID
		
		// update material status
		material.Status = entities.MaterialStatusInUse
		
		// save changes to the database
		if err := db.Save(slot).Error; err != nil {
			return err
		}
		if err := db.Save(material).Error; err != nil {
			return err
		}
		
		// create an operation record for the placement
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