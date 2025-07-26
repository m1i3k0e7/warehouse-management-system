const { Kafka } = require('kafkajs');
const logger = require('../utils/logger');

class KafkaController {
  constructor(realtimeService) {
    this.realtimeService = realtimeService;
    this.kafka = Kafka({
      clientId: 'realtime-service',
      brokers: process.env.KAFKA_BROKERS?.split(',') || ['localhost:9092'],
      retry: {
        initialRetryTime: 100,
        retries: 8
      }
    });
    
    this.consumer = this.kafka.consumer({ 
      groupId: 'realtime-service-group',
      sessionTimeout: 30000,
      heartbeatInterval: 3000
    });
    
    this.isRunning = false;
  }

  async start() {
    try {
      await this.consumer.connect();
      
      // 訂閱所有庫存相關事件
      await this.consumer.subscribe({ 
        topics: [
          'inventory_events',
          'shelf_events', 
          'system_alerts',
          'audit_logs'
        ],
        fromBeginning: false 
      });

      await this.consumer.run({
        eachMessage: async ({ topic, partition, message }) => {
          try {
            await this.handleMessage(topic, message);
          } catch (error) {
            logger.error('Error processing Kafka message:', error);
            // 在生產環境中，這裡可以將失敗的消息發送到死信隊列
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

    switch (topic) {
      case 'inventory_events':
        await this.handleInventoryEvent(eventData);
        break;
        
      case 'shelf_events':
        await this.handleShelfEvent(eventData);
        break;
        
      case 'system_alerts':
        await this.handleSystemAlert(eventData);
        break;
        
      case 'audit_logs':
        await this.handleAuditLog(eventData);
        break;
        
      default:
        logger.warn(`Unknown topic: ${topic}`);
    }
  }

  async handleInventoryEvent(eventData) {
    const { event_type, material_id, slot_id, shelf_id, operator_id } = eventData;
    
    switch (event_type) {
      case 'material.placed':
        await this.realtimeService.broadcastInventoryUpdate({
          shelf_id,
          event_type: 'material_placed',
          slot_id,
          material_id,
          operator_id,
          timestamp: eventData.timestamp
        });
        break;
        
      case 'material.removed':
        await this.realtimeService.broadcastInventoryUpdate({
          shelf_id,
          event_type: 'material_removed', 
          slot_id,
          material_id,
          operator_id,
          timestamp: eventData.timestamp
        });
        break;
        
      case 'material.moved':
        // 發送移除和放置兩個事件
        await this.realtimeService.broadcastInventoryUpdate({
          shelf_id,
          event_type: 'material_removed',
          slot_id: eventData.from_slot_id,
          material_id,
          operator_id,
          timestamp: eventData.timestamp
        });
        
        await this.realtimeService.broadcastInventoryUpdate({
          shelf_id,
          event_type: 'material_placed',
          slot_id,
          material_id,
          operator_id,
          timestamp: eventData.timestamp
        });
        break;
    }
  }

  async handleShelfEvent(eventData) {
    const { event_type, shelf_id } = eventData;
    
    switch (event_type) {
      case 'shelf.status_changed':
        // 廣播料架狀態變更
        this.realtimeService.broadcastShelfStatusChange({
          shelf_id,
          old_status: eventData.old_status,
          new_status: eventData.new_status,
          timestamp: eventData.timestamp
        });
        break;
        
      case 'shelf.health_alert':
        // 廣播健康告警
        this.realtimeService.broadcastHealthAlert(eventData);
        break;
    }
  }

  async handleSystemAlert(eventData) {
    const { alert_type, severity, message, metadata } = eventData;
    
    // 廣播系統告警到管理後台
    this.realtimeService.broadcastSystemAlert({
      type: alert_type,
      severity,
      message,
      metadata,
      timestamp: eventData.timestamp
    });
    
    // 如果是嚴重告警，也廣播到相關料架的工作人員
    if (severity === 'critical' && metadata.shelf_id) {
      this.realtimeService.broadcastToShelf(metadata.shelf_id, {
        type: 'critical_alert',
        message,
        timestamp: eventData.timestamp
      });
    }
  }

  async handleAuditLog(eventData) {
    // 將審計日誌發送到管理後台
    this.realtimeService.broadcastAuditLog(eventData);
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