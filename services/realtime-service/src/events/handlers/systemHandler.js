const { SHELF_EVENTS, SYSTEM_EVENTS } = require('../events/types');
const logger = require('../../utils/logger');
const { PHYSICAL_PLACEMENT_EVENTS } = require('../types');

class SystemEventHandler {
  constructor(realtimeService) {
    this.realtimeService = realtimeService;
  }

  async handle(eventType, eventData) {
    switch (eventType) {
      case SHELF_EVENTS.SHELF_STATUS_CHANGED:
        await this.realtimeService.broadcastShelfStatusChange(eventData);
        break;
      case SHELF_EVENTS.SHELF_HEALTH_ALERT:
        await this.realtimeService.broadcastHealthAlert(eventData);
        break;
      case SYSTEM_EVENTS.SYSTEM_ALERT:
        await this.realtimeService.broadcastSystemAlert(eventData);
        break;
      case SYSTEM_EVENTS.AUDIT_LOG:
        await this.realtimeService.broadcastAuditLog(eventData);
        break;
      default:
        logger.warn(`System event handler received unknown event type: ${eventType}`);
    }
  }

  async handlePlacementEvent(eventType, eventData) {
    let event = { 
      type: eventType,
      operation_id: eventData.operation_id,
      slot_id: eventData.slot_id,
      material_id: eventData.material_id,
      message: ""
    }

    switch (eventType) {
      case PHYSICAL_PLACEMENT_EVENTS.PLACEMENT_REQUESTED:
        event.message = `Please place material ${eventData.material_id} into slot ${eventData.slot_id} on shelf ${eventData.shelf_id}.`;
        break;
      case PHYSICAL_PLACEMENT_EVENTS.PLACEMENT_CONFIRMED:
        event.message = `Physical placement confirmed for operation ${eventData.operation_id}.`;
        break;
      case PHYSICAL_PLACEMENT_EVENTS.PLACEMENT_FAILED:
        event.message = `Physical placement failed for operation ${eventData.operation_id}. Slot rolled back.`;
        break;
      case PHYSICAL_PLACEMENT_EVENTS.PLACEMENT_UNPLANNED:
        event.message = `Unplanned physical placement detected for operation ${eventData.operation_id}.`;
        break;
    }

    switch (eventType) {
      case PHYSICAL_PLACEMENT_EVENTS.PLACEMENT_REQUESTED:
        await this.realtimeService.broadcastToShelf(eventData.shelf_id, event);
        break;
      case PHYSICAL_PLACEMENT_EVENTS.PLACEMENT_CONFIRMED:
      case PHYSICAL_PLACEMENT_EVENTS.PLACEMENT_FAILED:
      case PHYSICAL_PLACEMENT_EVENTS.PLACEMENT_UNPLANNED:
        await this.realtimeService.broadcastToShelf(eventData.shelf_id, event);
        await this.realtimeService.getShelfStatus(eventData.shelf_id).then(shelfStatus => {
          this.realtimeService.io.to(`shelf_${eventData.shelf_id}`).emit('system_event', { type: 'shelf_status', data: shelfStatus });
        });
        break;
      default:
        logger.warn(`System event handler received unknown placement event type: ${eventType}`);
    }
  }

  async handleRemovalEvent(eventType, eventData) {
    let event = { 
      type: eventType,
      operation_id: eventData.operation_id,
      slot_id: eventData.slot_id,
      material_id: eventData.material_id,
      message: ""
    }

    switch (eventType) {
      case PHYSICAL_REMOVAL_EVENTS.REMOVAL_REQUESTED:
        event.message = `Please remove material ${eventData.material_id} from slot ${eventData.slot_id} on shelf ${eventData.shelf_id}.`;
        break;
      case PHYSICAL_REMOVAL_EVENTS.REMOVAL_CONFIRMED:
        event.message = `Physical removal confirmed for operation ${eventData.operation_id}.`;
        break;
      case PHYSICAL_REMOVAL_EVENTS.REMOVAL_FAILED:
        event.message = `Physical removal failed for operation ${eventData.operation_id}. Slot rolled back.`;
        break;
      case PHYSICAL_REMOVAL_EVENTS.UNPLANNED_REMOVAL:
        event.message = `Unplanned physical removal detected for operation ${eventData.operation_id}.`;
        break;
    }

    switch (eventType) {
      case PHYSICAL_REMOVAL_EVENTS.REMOVAL_REQUESTED:
        await this.realtimeService.broadcastToShelf(eventData.shelf_id, event);
        break;
      case PHYSICAL_REMOVAL_EVENTS.REMOVAL_CONFIRMED:
      case PHYSICAL_REMOVAL_EVENTS.REMOVAL_FAILED:
      case PHYSICAL_REMOVAL_EVENTS.UNPLANNED_REMOVAL:
        await this.realtimeService.broadcastToShelf(eventData.shelf_id, event);
        await this.realtimeService.getShelfStatus(eventData.shelf_id).then(shelfStatus => {
          this.realtimeService.io.to(`shelf_${eventData.shelf_id}`).emit('system_event', { type: 'shelf_status', data: shelfStatus });
        });
        break;
      default:
        logger.warn(`System event handler received unknown removal event type: ${eventType}`);
    }
  }
}

module.exports = SystemEventHandler;
