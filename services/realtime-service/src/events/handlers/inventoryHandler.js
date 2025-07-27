const { INVENTORY_EVENTS } = require('../events/types');
const logger = require('../../utils/logger');

class InventoryEventHandler {
  constructor(realtimeService) {
    this.realtimeService = realtimeService;
  }

  async handle(eventType, eventData) {
    switch (eventType) {
      case INVENTORY_EVENTS.MATERIAL_PLACED:
      case INVENTORY_EVENTS.MATERIAL_REMOVED:
      case INVENTORY_EVENTS.MATERIAL_MOVED:
        await this.realtimeService.broadcastInventoryUpdate(eventData);
        break;
      default:
        logger.warn(`Inventory event handler received unknown event type: ${eventType}`);
    }
  }
}

module.exports = InventoryEventHandler;
