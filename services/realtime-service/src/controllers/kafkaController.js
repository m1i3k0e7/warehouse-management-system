const { Kafka } = require('kafkajs');
const logger = require('../utils/logger');
const { INVENTORY_EVENTS, SHELF_EVENTS, SYSTEM_EVENTS } = require('../events/types');
const InventoryEventHandler = require('../events/handlers/inventoryHandler');
const SystemEventHandler = require('../events/handlers/systemHandler');
const {config} = require('../config');

class KafkaController {
  constructor(realtimeService) {
    this.realtimeService = realtimeService;
    this.kafka = Kafka({
      clientId: config.kafka.clientId || 'realtime-service',
      brokers: config.kafka.brokers?.split(',') || ['localhost:9092'],
      retry: {
        initialRetryTime: config.kafka.retryInitialTime,
        retries: config.kafka.retryRetries
      }
    });
    
    this.consumer = this.kafka.consumer({ 
      groupId: config.kafka.groupId || 'realtime-service-group',
      sessionTimeout: 30000,
      heartbeatInterval: 3000
    });
    
    this.inventoryEventHandler = new InventoryEventHandler(realtimeService);
    this.systemEventHandler = new SystemEventHandler(realtimeService);
    this.isRunning = false;
  }

  async start() {
    try {
      await this.consumer.connect();
      
      // subscribe to multiple topics
      await this.consumer.subscribe({ 
        topics: Object.values(INVENTORY_EVENTS).concat(Object.values(SHELF_EVENTS), Object.values(SYSTEM_EVENTS)),
        fromBeginning: false 
      });

      await this.consumer.run({
        eachMessage: async ({ topic, partition, message }) => {
          try {
            await this.handleMessage(topic, message);
          } catch (error) {
            logger.error('Error processing Kafka message:', error);
            // in production, you might want to handle retries or dead-letter queues
          }
        },
      });

      this.isRunning = true;
      logger.info('Kafka consumer started successfully');
    } catch (error) {
      logger.error('Failed to start Kafka consumer:', error);
      throw error;
    }
  }

  async handleMessage(topic, message) {
    const eventData = JSON.parse(message.value.toString());
    const eventType = eventData.event_type || message.key?.toString();
    
    logger.info(`Processing Kafka message`, { topic, eventType });

    switch (eventType) {
      case INVENTORY_EVENTS.MATERIAL_PLACED:
      case INVENTORY_EVENTS.MATERIAL_REMOVED:
      case INVENTORY_EVENTS.MATERIAL_MOVED:
        await this.inventoryEventHandler.handle(eventType, eventData);
        break;
      case SHELF_EVENTS.SHELF_STATUS_CHANGED:
      case SHELF_EVENTS.SHELF_HEALTH_ALERT:
      case SYSTEM_EVENTS.SYSTEM_ALERT:
      case SYSTEM_EVENTS.AUDIT_LOG:
        await this.systemEventHandler.handle(eventType, eventData);
        break;
      default:
        logger.warn(`Unknown event type received: ${eventType} on topic ${topic}`);
    }
  }

  async stop() {
    if (this.isRunning) {
      try {
        await this.consumer.disconnect();
        this.isRunning = false;
        logger.info('Kafka consumer stopped');
      } catch (error) {
        logger.error('Error stopping Kafka consumer:', error);
      }
    }
  }
}

module.exports = KafkaController;