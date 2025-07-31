package services

const (
	// Inventory Events
	EventTypeMaterialPlaced = "material.placed"
	EventTypeMaterialRemoved = "material.removed"
	EventTypeMaterialMoved = "material.moved"

	// Physical Placement Events
	EventTypeMaterialDetected = "material.detected" // Raw event from physical sensor
	EventTypeUnplannedPlacement = "unplanned.placement" // Event for detected but unplanned placement
	EventTypePhysicalPlacementRequested = "physical.placement.requested" // Event for requested physical placement
	EventTypePhysicalPlacementConfirmed = "physical.placement.confirmed" // Event for confirmed physical placement
	EventTypePhysicalPlacementFailed = "physical.placement.failed" // Event for failed physical placement

	// Physical Removal Events
	EventTypeMaterialRemovedFromShelf = "material.removed_from_shelf" // Event for material removed from shelf
	EventTypeUnplannedRemoval = "unplanned.removal" // Event for detected but unplanned removal
	EventTypePhysicalRemovalRequested = "physical.removal.requested" // Event for requested physical removal
	EventTypePhysicalRemovalConfirmed = "physical.removal.confirmed" // Event for confirmed physical removal
	EventTypePhysicalRemovalFailed = "physical.removal.failed" // Event for failed physical removal

	// Shelf Events
	EventTypeShelfStatusChanged = "shelf.status_changed"
	EventTypeShelfHealthAlert = "shelf.health_alert"

	// System Events
	EventTypeSystemAlert = "system.alert"
	EventTypeAuditLog = "audit.log"
)
