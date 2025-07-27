const { SHELF_EVENTS, SYSTEM_EVENTS } = require('../events/types');
const logger = require('../../utils/logger');

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
}

module.exports = SystemEventHandler;
