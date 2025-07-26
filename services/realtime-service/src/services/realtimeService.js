const redis = require('../utils/redis');
const logger = require('../utils/logger');
const InventoryAPIService = require('./inventoryAPIService');

class RealtimeService {
  constructor(io) {
    this.io = io;
    this.inventoryAPIService = new InventoryAPIService();
  }

  // 廣播庫存更新事件
  async broadcastInventoryUpdate(event) {
    const { shelf_id, event_type, slot_id, material_id } = event;
    
    try {
      // 廣播給該料架的所有工作人員
      this.io.to(`shelf_${shelf_id}`).emit('inventory_update', {
        type: event_type,
        data: {
          shelfId: shelf_id,
          slotId: slot_id,
          materialId: material_id,
          timestamp: new Date().toISOString()
        }
      });

      // 廣播給管理後台
      this.io.to('admin_dashboard').emit('global_update', {
        type: 'inventory_change',
        data: event
      });

      // 更新實時統計
      await this.updateRealtimeStats(shelf_id, event_type);
      
      logger.info(`Broadcasted inventory update`, { 
        shelfId: shelf_id, 
        eventType: event_type 
      });
    } catch (error) {
      logger.error('Failed to broadcast inventory update:', error);
    }
  }

  // 處理客戶端加入料架房間
  async joinShelfRoom(socket, shelfId, operatorId) {
    try {
      socket.join(`shelf_${shelfId}`);
      
      // 記錄會話信息到 Redis
      const sessionData = {
        shelfId,
        operatorId,
        joinedAt: new Date().toISOString()
      };
      await redis.set(`session:${socket.id}`, JSON.stringify(sessionData), 'EX', 3600); // 設置會話過期時間為1小時

      // 發送當前料架狀態
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

  // 處理實時操作請求
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

  // 處理操作請求的內部邏輯
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

  // 處理客戶端斷開連接
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

  // 獲取料架狀態
  async getShelfStatus(shelfId) {
    try {
      const cachedStatus = await redis.get(`shelf_status:${shelfId}`);
      if (cachedStatus) {
        return JSON.parse(cachedStatus);
      }
      
      // 如果緩存中沒有，調用庫存服務 API
      const apiStatus = await this.inventoryAPIService.getShelfStatus(shelfId);
      await redis.set(`shelf_status:${shelfId}`, JSON.stringify(apiStatus), 'EX', 600); // 緩存10分鐘
      return apiStatus;
    } catch (error) {
      logger.error('Failed to get shelf status:', error);
      return { shelfId, status: 'error' };
    }
  }

  // 更新實時統計
  async updateRealtimeStats(shelfId, eventType) {
    const key = `stats:${shelfId}:${new Date().toISOString().slice(0, 10)}`;
    await redis.hincrby(key, eventType, 1);
    await redis.expire(key, 86400 * 7); // 保存7天
  }

  

  async broadcastShelfStatusChange(data) {
    const { shelf_id, old_status, new_status, timestamp } = data;
    
    try {
      // 廣播給該料架的工作人員
      this.io.to(`shelf_${shelf_id}`).emit('shelf_status_changed', {
        shelfId: shelf_id,
        oldStatus: old_status,
        newStatus: new_status,
        timestamp
      });

      // 廣播給管理後台
      this.io.to('admin_dashboard').emit('shelf_status_update', data);
      
      logger.info(`Broadcasted shelf status change`, { shelf_id, old_status, new_status });
    } catch (error) {
      logger.error('Failed to broadcast shelf status change:', error);
    }
  }

  // 廣播健康告警
  async broadcastHealthAlert(alertData) {
    try {
      const { shelf_id, health_score, message, severity } = alertData;
      
      // 發送給相關料架的工作人員
      this.io.to(`shelf_${shelf_id}`).emit('health_alert', {
        type: 'shelf_health',
        shelfId: shelf_id,
        healthScore: health_score,
        message,
        severity,
        timestamp: alertData.timestamp
      });

      // 發送給管理後台
      this.io.to('admin_dashboard').emit('health_alert', alertData);
      
      logger.info(`Broadcasted health alert`, { shelf_id, severity });
    } catch (error) {
      logger.error('Failed to broadcast health alert:', error);
    }
  }

  // 廣播系統告警
  async broadcastSystemAlert(alertData) {
    try {
      // 主要發送給管理後台
      this.io.to('admin_dashboard').emit('system_alert', alertData);
      
      // 如果是緊急告警，廣播給所有連接的客戶端
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

  // 廣播審計日誌
  async broadcastAuditLog(logData) {
    try {
      // 只發送給管理後台
      this.io.to('admin_dashboard').emit('audit_log', logData);
    } catch (error) {
      logger.error('Failed to broadcast audit log:', error);
    }
  }

  // 向特定料架廣播消息
  async broadcastToShelf(shelfId, message) {
    try {
      this.io.to(`shelf_${shelfId}`).emit('shelf_message', message);
    } catch (error) {
      logger.error('Failed to broadcast to shelf:', error);
    }
  }
}

module.exports = RealtimeService;