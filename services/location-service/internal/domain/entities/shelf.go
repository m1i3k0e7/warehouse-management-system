package entities

// Point represents a 3D coordinate in the warehouse.
type Point struct {
	X int
	Y int
	Z int
}

// SlotStatus represents the state of a shelf slot.
type SlotStatus string

const (
	StatusEmpty    SlotStatus = "EMPTY"
	StatusOccupied SlotStatus = "OCCUPIED"
	StatusReserved SlotStatus = "RESERVED"
	StatusDisabled SlotStatus = "DISABLED"
)

// Slot represents a single storage unit on a shelf.
type Slot struct {
	ID         string
	Position   Point
	Status     SlotStatus
	MaterialID string // Foreign key to material in inventory-service
}

// Shelf represents a physical shelf in the warehouse.
type Shelf struct {
	ID       string
	ZoneID   string
	Position Point
	Rows     int
	Columns  int
	Slots    []Slot
}