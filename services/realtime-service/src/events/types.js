const INVENTORY_EVENTS = {
  MATERIAL_PLACED: 'material.placed',
  MATERIAL_REMOVED: 'material.removed',
  MATERIAL_MOVED: 'material.moved',
};

const PHYSICAL_PLACEMENT_EVENTS = {
  PHYSICAL_PLACEMENT_REQUESTED: 'physical.placement.requested',
  PHYSICAL_PLACEMENT_CONFIRMED: 'physical.placement.confirmed',
  PHYSICAL_PLACEMENT_FAILED: 'physical.placement.failed',
  PHYSICAL_PLACEMENT_UNPLANNED: 'physical.placement.unplanned',
};

const PHYSICAL_REMOVAL_EVENTS = {
  PHYSICAL_REMOVAL_REQUESTED: 'physical.removal.requested',
  PHYSICAL_REMOVAL_CONFIRMED: 'physical.removal.confirmed',
  PHYSICAL_REMOVAL_FAILED: 'physical.removal.failed',
  PHYSICAL_REMOVAL_UNPLANNED: 'physical.removal.unplanned',
};

const PHYSICAL_MOVE_EVENTS = {
  PHYSICAL_MOVE_REQUESTED: 'physical.move.requested',
  PHYSICAL_MOVE_CONFIRMED: 'physical.move.confirmed',
  PHYSICAL_MOVE_FAILED: 'physical.move.failed',
  PHYSICAL_MOVE_UNPLANNED: 'physical.move.unplanned',
};

const SLOTS_RESERVATION_EVENTS = {
  SLOTS_RESERVED: 'slots.reserved',
  SLOTS_RELEASED: 'slots.released',
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
