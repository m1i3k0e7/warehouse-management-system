package services

const (
	// Inventory Events
	EventTypeMaterialPlaced = "material.placed"
	EventTypeMaterialRemoved = "material.removed"
	EventTypeMaterialMoved = "material.moved"

	// Shelf Events
	EventTypeShelfStatusChanged = "shelf.status_changed"
	EventTypeShelfHealthAlert = "shelf.health_alert"

	// System Events
	EventTypeSystemAlert = "system.alert"
	EventTypeAuditLog = "audit.log"
)
