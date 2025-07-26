const redis = require('../utils/redis');
const logger = require('../utils/logger');

class RealtimeService {
  constructor(io) {
    this.io = io;
    this.rooms = new Map(); // 房間管理
    this.clientSessions = new Map(); // 客戶端會話
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
      
      // 記錄會話信息
      this.clientSessions.set(socket.id, {
        shelfId,
        operatorId,
        joinedAt: new Date()
      });

      // 發送當前料架狀態
      const shelfStatus = await this.getShelfStatus(shelfId);
      socket.emit('shelf_status', shelfStatus);

      // 更新房間統計
      this.updateRoomStats(`shelf_${shelfId}`, 'join');
      
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
    const session = this.clientSessions.get(socket.id);
    if (!session) {
      socket.emit('operation_response', { 
        success: false, 
        error: 'Session not found' 
      });
      return;
    }

    try {
      // 這裡可以調用其他服務的 API 來處理操作
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

  // 處理客戶端斷開連接
  handleDisconnect(socket) {
    const session = this.clientSessions.get(socket.id);
    if (session) {
      this.updateRoomStats(`shelf_${session.shelfId}`, 'leave');
      this.clientSessions.delete(socket.id);
      
      logger.info(`Client disconnected`, { 
        socketId: socket.id, 
        shelfId: session.shelfId 
      });
    }
  }

  // 獲取料架狀態
  async getShelfStatus(shelfId) {
    try {
      const cachedStatus = await redis.hgetall(`shelf:${shelfId}`);
      if (cachedStatus && cachedStatus.data) {
        return JSON.parse(cachedStatus.data);
      }
      
      // 如果緩存中沒有，可以調用庫存服務 API
      return { shelfId, status: 'unknown' };
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

  // 更新房間統計
  updateRoomStats(roomName, action) {
    if (!this.rooms.has(roomName)) {
      this.rooms.set(roomName, { activeConnections: 0 });
    }
    
    const room = this.rooms.get(roomName);
    if (action === 'join') {
      room.activeConnections++;
    } else if (action === 'leave') {
      room.activeConnections = Math.max(0, room.activeConnections - 1);
    }
  }
}

module.exports = RealtimeService;