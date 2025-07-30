package services

const (
	// Inventory Events
	EventTypeMaterialPlaced = "material.placed"
	EventTypeMaterialRemoved = "material.removed"
	EventTypeMaterialMoved = "material.moved"

	// Physical Placement Events
	EventTypeMaterialDetected = "material.detected" // Raw event from physical sensor
	EventTypeUnplannedPlacement = "unplanned.placement" // Event for detected but unplanned placement

	// Shelf Events
	EventTypeShelfStatusChanged = "shelf.status_changed"
	EventTypeShelfHealthAlert = "shelf.health_alert"

	// System Events
	EventTypeSystemAlert = "system.alert"
	EventTypeAuditLog = "audit.log"
)
