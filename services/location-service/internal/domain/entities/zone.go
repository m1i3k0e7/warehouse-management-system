package entities

// Zone represents a logical area in the warehouse, like "Receiving", "Packing", or "High-Value Storage".
type Zone struct {
	ID             string
	Name           string
	BoundaryPoints []Point // Defines the geographical area of the zone
}