const redis = require('../utils/redis');
const logger = require('../utils/logger');
const InventoryAPIService = require('./inventoryAPIService');

class RealtimeService {
  constructor(io, notificationService, roomService) {
    this.io = io;
    this.inventoryAPIService = new InventoryAPIService();
    this.notificationService = notificationService;
    this.roomService = roomService;
  }

  async broadcastInventoryUpdate(event) {
    const { shelf_id, event_type, slot_id, material_id } = event;
    
    try {
      // broadcast to specific shelf room
      this.roomService.broadcastToRoom(`shelf_${shelf_id}`, 'inventory_update', {
        type: event_type,
        data: {
          shelfId: shelf_id,
          slotId: slot_id,
          materialId: material_id,
          timestamp: new Date().toISOString()
        }
      });

      // broadcast to admin dashboard
      this.roomService.broadcastToRoom('admin_dashboard', 'global_update', {
        type: 'inventory_change',
        data: event
      });

      // update realtime statistics
      await this.updateRealtimeStats(shelf_id, event_type);

      // update shelf status cache
      const updatedShelfStatus = await this.inventoryAPIService.getShelfStatus(shelf_id);
      await redis.set(`shelf_status:${shelf_id}`, JSON.stringify(updatedShelfStatus), 'EX', 600); // expires in 600 seconds (10 minutes)
      
      logger.info(`Broadcasted inventory update and updated shelf status cache`, { 
        shelfId: shelf_id, 
        eventType: event_type 
      });
    } catch (error) {
      logger.error('Failed to broadcast inventory update or update cache:', error);
    }
  }

  async joinShelfRoom(socket, shelfId, operatorId) {
    try {
      this.roomService.joinRoom(socket, `shelf_${shelfId}`);
      
      // record session data in Redis
      const sessionData = {
        shelfId,
        operatorId,
        joinedAt: new Date().toISOString()
      };
      await redis.set(`session:${socket.id}`, JSON.stringify(sessionData), 'EX', 3600); // expires in 3600 seconds (1 hour)

      // send current shelf status to the client
      const shelfStatus = await this.getShelfStatus(shelfId);
      socket.emit('shelf_status', shelfStatus);
      
      logger.info(`Client joined shelf room`, { 
        socketId: socket.id, 
        shelfId, 
        operatorId 
      });
    } catch (error) {
      logger.error('Failed to join shelf room:', error);
      socket.emit('error', { message: 'Failed to join shelf room' });
    }
  }

  async handleOperationRequest(socket, data) {
    const sessionData = await redis.get(`session:${socket.id}`);
    if (!sessionData) {
      socket.emit('operation_response', {
        success: false,
        error: 'Session not found' 
      });
      return;
    }
    const session = JSON.parse(sessionData);

    try {
      const result = await this.processOperation({
        ...data,
        operatorId: session.operatorId,
        shelfId: session.shelfId
      });

      socket.emit('operation_response', {
        success: true, 
        data: result 
      });
    } catch (error) {
      logger.error('Operation request failed:', error);
      socket.emit('operation_response', { 
        success: false, 
        error: error.message 
      });
    }
  }

  async processOperation(operationData) {
    const { type, materialBarcode, slotId, fromSlotId, toSlotId, operatorId, shelfId, reason, duration, purpose } = operationData;

    switch (type) {
      case 'place_material':
        return this.inventoryAPIService.placeMaterial({
          materialBarcode,
          slotId,
          operatorId,
        });
      case 'remove_material':
        return this.inventoryAPIService.removeMaterial({
          slotId,
          operatorId,
          reason,
        });
      case 'move_material':
        return this.inventoryAPIService.moveMaterial({
          fromSlotId,
          toSlotId,
          operatorId,
          reason,
        });
      // Add more operation types as needed
      default:
        throw new Error(`Unknown operation type: ${type}`);
    }
  }

  async handleDisconnect(socket) {
    const sessionData = await redis.get(`session:${socket.id}`);
    if (sessionData) {
      const session = JSON.parse(sessionData);
      await redis.del(`session:${socket.id}`);
      
      logger.info(`Client disconnected`, { 
        socketId: socket.id, 
        shelfId: session.shelfId 
      });
    }
  }

  async getShelfStatus(shelfId) {
    try {
      const cachedStatus = await redis.get(`shelf_status:${shelfId}`);
      if (cachedStatus) {
        return JSON.parse(cachedStatus);
      }
      
      // call the inventory API to get the latest status if not cached
      const apiStatus = await this.inventoryAPIService.getShelfStatus(shelfId);
      await redis.set(`shelf_status:${shelfId}`, JSON.stringify(apiStatus), 'EX', 600); // expires in 600 seconds (10 minutes)
      return apiStatus;
    } catch (error) {
      logger.error('Failed to get shelf status:', error);
      return { shelfId, status: 'error' };
    }
  }

  async updateRealtimeStats(shelfId, eventType) {
    const key = `stats:${shelfId}:${new Date().toISOString().slice(0, 10)}`;
    await redis.hincrby(key, eventType, 1);
    await redis.expire(key, 86400 * 7); // set key to expire in 7 days
  }

  async broadcastShelfStatusChange(data) {
    const { shelf_id, old_status, new_status, timestamp } = data;
    
    try {
      // broadcast to specific shelf room
      this.roomService.broadcastToRoom(`shelf_${shelf_id}`, 'shelf_status_changed', {
        shelfId: shelf_id,
        oldStatus: old_status,
        newStatus: new_status,
        timestamp
      });

      // broadcast to admin dashboard
      this.roomService.broadcastToRoom('admin_dashboard', 'shelf_status_update', data);
      
      logger.info(`Broadcasted shelf status change`, { shelf_id, old_status, new_status });
    } catch (error) {
      logger.error('Failed to broadcast shelf status change:', error);
    }
  }

  async broadcastHealthAlert(alertData) {
    try {
      const { shelf_id, health_score, message, severity } = alertData;
      
      // broadcast to specific shelf room
      this.roomService.broadcastToRoom(`shelf_${shelf_id}`, 'health_alert', {
        type: 'shelf_health',
        shelfId: shelf_id,
        healthScore: health_score,
        message,
        severity,
        timestamp: alertData.timestamp
      });

      // broadcast to admin dashboard
      this.roomService.broadcastToRoom('admin_dashboard', 'health_alert', alertData);
      
      logger.info(`Broadcasted health alert`, { shelf_id, severity });
    } catch (error) {
      logger.error('Failed to broadcast health alert:', error);
    }
  }

  async broadcastSystemAlert(alertData) {
    try {
      // broadcast to admin dashboard
      this.roomService.broadcastToRoom('admin_dashboard', 'system_alert', alertData);
      
      // broadcast critical alerts to all connected clients if severity is critical
      if (alertData.severity === 'critical') {
        this.io.emit('critical_system_alert', {
          message: alertData.message,
          timestamp: alertData.timestamp
        });
      }
      
      logger.info(`Broadcasted system alert`, { type: alertData.type, severity: alertData.severity });
    } catch (error) {
      logger.error('Failed to broadcast system alert:', error);
    }
  }

  async broadcastAuditLog(logData) {
    try {
      // broadcast to admin dashboard
      this.roomService.broadcastToRoom('admin_dashboard', 'audit_log', logData);
    } catch (error) {
      logger.error('Failed to broadcast audit log:', error);
    }
  }

  // broadcast a message to a specific shelf
  async broadcastToShelf(shelfId, message) {
    try {
      this.roomService.broadcastToRoom(`shelf_${shelfId}`, 'shelf_message', message);
    } catch (error) {
      logger.error('Failed to broadcast to shelf:', error);
    }
  }
}

module.exports = RealtimeService;