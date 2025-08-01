package entities

// Path represents a sequence of points from a start to an end location.
type Path struct {
	Points   []Point
	Distance float64
}