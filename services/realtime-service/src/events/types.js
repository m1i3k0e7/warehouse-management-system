const INVENTORY_EVENTS = {
  MATERIAL_PLACED: 'material.placed',
  MATERIAL_REMOVED: 'material.removed',
  MATERIAL_MOVED: 'material.moved',
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
  SHELF_EVENTS,
  SYSTEM_EVENTS,
};
