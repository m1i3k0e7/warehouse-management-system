const INVENTORY_EVENTS = {
  MATERIAL_PLACED: 'material.placed',
  MATERIAL_REMOVED: 'material.removed',
  MATERIAL_MOVED: 'material.moved',
};

const PHYSICAL_PLACEMENT_EVENTS = {
  PHYSICAL_PLACEMENT_REQUESTED: 'physical.placement.requested',
  PHYSICAL_PLACEMENT_CONFIRMED: 'physical.placement.confirmed',
  PHYSICAL_PLACEMENT_FAILED: 'physical.placement.failed',
};

const SHELF_EVENTS = {
  SHELF_STATUS_CHANGED: 'shelf.status_changed',
  SHELF_HEALTH_ALERT: 'shelf.health_alert',
};

const SYSTEM_EVENTS = {
  SYSTEM_ALERT: 'system.alert',
  AUDIT_LOG: 'audit.log',
};

module.exports = {
  INVENTORY_EVENTS,
  PHYSICAL_PLACEMENT_EVENTS,
  SHELF_EVENTS,
  SYSTEM_EVENTS,
};
