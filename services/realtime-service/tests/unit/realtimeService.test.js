const RealtimeService = require('../../src/services/realtimeService');
const InventoryAPIService = require('../../src/services/inventoryAPIService');
const NotificationService = require('../../src/services/notificationService');
const RoomService = require('../../src/services/roomService');
const redis = require('../../src/utils/redis');
const logger = require('../../src/utils/logger');

// Mock dependencies
jest.mock('../../src/services/inventoryAPIService');
jest.mock('../../src/services/notificationService');
jest.mock('../../src/services/roomService');
jest.mock('../../src/utils/redis');
jest.mock('../../src/utils/logger');

describe('RealtimeService', () => {
  let ioMock;
  let realtimeService;
  let socketMock;

  beforeEach(() => {
    // Reset mocks before each test
    jest.clearAllMocks();

    // Mock Socket.IO server instance
    ioMock = {
      to: jest.fn().mockReturnThis(),
      emit: jest.fn(),
    };

    // Mock a socket connection
    socketMock = {
      id: 'test-socket-id',
      join: jest.fn(),
      emit: jest.fn(),
      handshake: { address: '127.0.0.1' },
    };

    // Instantiate RealtimeService with mocked dependencies
    realtimeService = new RealtimeService(ioMock, new NotificationService(), new RoomService(ioMock));
  });

  describe('broadcastInventoryUpdate', () => {
    test('should broadcast inventory update and update shelf status cache', async () => {
      const event = {
        shelf_id: 'shelf1',
        event_type: 'material.placed',
        slot_id: 'slot1',
        material_id: 'material1',
      };
      const updatedShelfStatus = { shelfId: 'shelf1', totalSlots: 100 };

      InventoryAPIService.mock.instances[0].getShelfStatus.mockResolvedValue(updatedShelfStatus);
      redis.set.mockResolvedValue('OK');

      await realtimeService.broadcastInventoryUpdate(event);

      // Verify broadcast to shelf room
      expect(realtimeService.roomService.broadcastToRoom).toHaveBeenCalledWith(
        `shelf_${event.shelf_id}`,
        'inventory_update',
        expect.objectContaining({ type: event.event_type })
      );

      // Verify broadcast to admin dashboard
      expect(realtimeService.roomService.broadcastToRoom).toHaveBeenCalledWith(
        'admin_dashboard',
        'global_update',
        expect.objectContaining({ type: 'inventory_change' })
      );

      // Verify shelf status cache update
      expect(InventoryAPIService.mock.instances[0].getShelfStatus).toHaveBeenCalledWith(event.shelf_id);
      expect(redis.set).toHaveBeenCalledWith(
        `shelf_status:${event.shelf_id}`,
        JSON.stringify(updatedShelfStatus),
        'EX',
        600
      );
      expect(logger.info).toHaveBeenCalled();
    });

    test('should log error if broadcasting fails', async () => {
      const event = {
        shelf_id: 'shelf1',
        event_type: 'material.placed',
        slot_id: 'slot1',
        material_id: 'material1',
      };
      const error = new Error('Broadcast failed');

      realtimeService.roomService.broadcastToRoom.mockImplementation(() => {
        throw error;
      });

      await realtimeService.broadcastInventoryUpdate(event);

      expect(logger.error).toHaveBeenCalledWith('Failed to broadcast inventory update or update cache:', error);
    });
  });

  describe('joinShelfRoom', () => {
    test('should join room, set session, and emit shelf status', async () => {
      const shelfId = 'shelf1';
      const operatorId = 'op1';
      const shelfStatus = { shelfId: 'shelf1', status: 'online' };

      redis.set.mockResolvedValue('OK');
      InventoryAPIService.mock.instances[0].getShelfStatus.mockResolvedValue(shelfStatus);

      await realtimeService.joinShelfRoom(socketMock, shelfId, operatorId);

      expect(realtimeService.roomService.joinRoom).toHaveBeenCalledWith(socketMock, `shelf_${shelfId}`);
      expect(redis.set).toHaveBeenCalledWith(
        `session:${socketMock.id}`,
        JSON.stringify({ shelfId, operatorId, joinedAt: expect.any(String) }),
        'EX',
        3600
      );
      expect(InventoryAPIService.mock.instances[0].getShelfStatus).toHaveBeenCalledWith(shelfId);
      expect(socketMock.emit).toHaveBeenCalledWith('shelf_status', shelfStatus);
      expect(logger.info).toHaveBeenCalled();
    });

    test('should emit error if joining room fails', async () => {
      const shelfId = 'shelf1';
      const operatorId = 'op1';
      const error = new Error('Join room failed');

      realtimeService.roomService.joinRoom.mockImplementation(() => {
        throw error;
      });

      await realtimeService.joinShelfRoom(socketMock, shelfId, operatorId);

      expect(logger.error).toHaveBeenCalledWith('Failed to join shelf room:', error);
      expect(socketMock.emit).toHaveBeenCalledWith('error', { message: 'Failed to join shelf room' });
    });
  });

  describe('handleOperationRequest', () => {
    test('should process operation and emit success response', async () => {
      const data = { type: 'place_material', materialBarcode: 'mat1', slotId: 'slot1' };
      const session = { operatorId: 'op1', shelfId: 'shelf1' };
      const processResult = { message: 'Material placed' };

      redis.get.mockResolvedValue(JSON.stringify(session));
      realtimeService.processOperation = jest.fn().mockResolvedValue(processResult);

      await realtimeService.handleOperationRequest(socketMock, data);

      expect(redis.get).toHaveBeenCalledWith(`session:${socketMock.id}`);
      expect(realtimeService.processOperation).toHaveBeenCalledWith({
        ...data,
        operatorId: session.operatorId,
        shelfId: session.shelfId,
      });
      expect(socketMock.emit).toHaveBeenCalledWith('operation_response', { success: true, data: processResult });
    });

    test('should emit error if session not found', async () => {
      const data = { type: 'place_material' };
      redis.get.mockResolvedValue(null);

      await realtimeService.handleOperationRequest(socketMock, data);

      expect(redis.get).toHaveBeenCalledWith(`session:${socketMock.id}`);
      expect(socketMock.emit).toHaveBeenCalledWith('operation_response', { success: false, error: 'Session not found' });
    });

    test('should emit error if operation processing fails', async () => {
      const data = { type: 'place_material' };
      const session = { operatorId: 'op1', shelfId: 'shelf1' };
      const error = new Error('Operation failed');

      redis.get.mockResolvedValue(JSON.stringify(session));
      realtimeService.processOperation = jest.fn().mockRejectedValue(error);

      await realtimeService.handleOperationRequest(socketMock, data);

      expect(logger.error).toHaveBeenCalledWith('Operation request failed:', error);
      expect(socketMock.emit).toHaveBeenCalledWith('operation_response', { success: false, error: error.message });
    });
  });

  describe('handleDisconnect', () => {
    test('should delete session from redis and log disconnection', async () => {
      const session = { shelfId: 'shelf1', operatorId: 'op1' };
      redis.get.mockResolvedValue(JSON.stringify(session));
      redis.del.mockResolvedValue(1);

      await realtimeService.handleDisconnect(socketMock);

      expect(redis.get).toHaveBeenCalledWith(`session:${socketMock.id}`);
      expect(redis.del).toHaveBeenCalledWith(`session:${socketMock.id}`);
      expect(logger.info).toHaveBeenCalledWith(
        `Client disconnected`,
        expect.objectContaining({ socketId: socketMock.id, shelfId: session.shelfId })
      );
    });

    test('should not delete session if not found', async () => {
      redis.get.mockResolvedValue(null);

      await realtimeService.handleDisconnect(socketMock);

      expect(redis.get).toHaveBeenCalledWith(`session:${socketMock.id}`);
      expect(redis.del).not.toHaveBeenCalled();
      expect(logger.info).not.toHaveBeenCalled();
    });
  });

  describe('getShelfStatus', () => {
    test('should return cached status if available', async () => {
      const cachedStatus = { shelfId: 'shelf1', status: 'online' };
      redis.get.mockResolvedValue(JSON.stringify(cachedStatus));

      const result = await realtimeService.getShelfStatus('shelf1');

      expect(redis.get).toHaveBeenCalledWith('shelf_status:shelf1');
      expect(result).toEqual(cachedStatus);
      expect(InventoryAPIService.mock.instances[0].getShelfStatus).not.toHaveBeenCalled();
    });

    test('should fetch from API and cache if not available', async () => {
      const apiStatus = { shelfId: 'shelf1', status: 'offline' };
      redis.get.mockResolvedValue(null);
      InventoryAPIService.mock.instances[0].getShelfStatus.mockResolvedValue(apiStatus);
      redis.set.mockResolvedValue('OK');

      const result = await realtimeService.getShelfStatus('shelf1');

      expect(redis.get).toHaveBeenCalledWith('shelf_status:shelf1');
      expect(InventoryAPIService.mock.instances[0].getShelfStatus).toHaveBeenCalledWith('shelf1');
      expect(redis.set).toHaveBeenCalledWith(
        'shelf_status:shelf1',
        JSON.stringify(apiStatus),
        'EX',
        600
      );
      expect(result).toEqual(apiStatus);
    });

    test('should log error if fetching status fails', async () => {
      const error = new Error('API error');
      redis.get.mockResolvedValue(null);
      InventoryAPIService.mock.instances[0].getShelfStatus.mockRejectedValue(error);

      const result = await realtimeService.getShelfStatus('shelf1');

      expect(logger.error).toHaveBeenCalledWith('Failed to get shelf status:', error);
      expect(result).toEqual({ shelfId: 'shelf1', status: 'error' });
    });
  });

  describe('updateRealtimeStats', () => {
    test('should increment and set expiry for stats', async () => {
      const shelfId = 'shelf1';
      const eventType = 'material.placed';
      const expectedKey = `stats:${shelfId}:${new Date().toISOString().slice(0, 10)}`;

      redis.hincrby.mockResolvedValue(1);
      redis.expire.mockResolvedValue(1);

      await realtimeService.updateRealtimeStats(shelfId, eventType);

      expect(redis.hincrby).toHaveBeenCalledWith(expectedKey, eventType, 1);
      expect(redis.expire).toHaveBeenCalledWith(expectedKey, 86400 * 7);
    });
  });

  describe('broadcastShelfStatusChange', () => {
    test('should broadcast shelf status change to room and admin dashboard', async () => {
      const data = { shelf_id: 'shelf1', old_status: 'offline', new_status: 'online', timestamp: '2023-01-01' };

      await realtimeService.broadcastShelfStatusChange(data);

      expect(realtimeService.roomService.broadcastToRoom).toHaveBeenCalledWith(
        `shelf_${data.shelf_id}`,
        'shelf_status_changed',
        expect.objectContaining({ shelfId: data.shelf_id })
      );
      expect(realtimeService.roomService.broadcastToRoom).toHaveBeenCalledWith(
        'admin_dashboard',
        'shelf_status_update',
        data
      );
      expect(logger.info).toHaveBeenCalled();
    });
  });

  describe('broadcastHealthAlert', () => {
    test('should broadcast health alert to room and admin dashboard', async () => {
      const alertData = { shelf_id: 'shelf1', health_score: 80, message: 'Low health', severity: 'medium', timestamp: '2023-01-01' };

      await realtimeService.broadcastHealthAlert(alertData);

      expect(realtimeService.roomService.broadcastToRoom).toHaveBeenCalledWith(
        `shelf_${alertData.shelf_id}`,
        'health_alert',
        expect.objectContaining({ shelfId: alertData.shelf_id })
      );
      expect(realtimeService.roomService.broadcastToRoom).toHaveBeenCalledWith(
        'admin_dashboard',
        'health_alert',
        alertData
      );
      expect(logger.info).toHaveBeenCalled();
    });
  });

  describe('broadcastSystemAlert', () => {
    test('should broadcast system alert to admin dashboard', async () => {
      const alertData = { type: 'sensor_error', severity: 'low', message: 'Sensor issue', timestamp: '2023-01-01' };

      await realtimeService.broadcastSystemAlert(alertData);

      expect(realtimeService.roomService.broadcastToRoom).toHaveBeenCalledWith(
        'admin_dashboard',
        'system_alert',
        alertData
      );
      expect(ioMock.emit).not.toHaveBeenCalled(); // Not critical, so not to all
      expect(logger.info).toHaveBeenCalled();
    });

    test('should broadcast critical system alert to all clients', async () => {
      const alertData = { type: 'power_failure', severity: 'critical', message: 'Power outage', timestamp: '2023-01-01' };

      await realtimeService.broadcastSystemAlert(alertData);

      expect(realtimeService.roomService.broadcastToRoom).toHaveBeenCalledWith(
        'admin_dashboard',
        'system_alert',
        alertData
      );
      expect(ioMock.emit).toHaveBeenCalledWith(
        'critical_system_alert',
        expect.objectContaining({ message: alertData.message })
      );
      expect(logger.info).toHaveBeenCalled();
    });
  });

  describe('broadcastAuditLog', () => {
    test('should broadcast audit log to admin dashboard', async () => {
      const logData = { action: 'login', operatorId: 'op1' };

      await realtimeService.broadcastAuditLog(logData);

      expect(realtimeService.roomService.broadcastToRoom).toHaveBeenCalledWith(
        'admin_dashboard',
        'audit_log',
        logData
      );
    });
  });

  describe('broadcastToShelf', () => {
    test('should broadcast message to specific shelf room', async () => {
      const shelfId = 'shelf1';
      const message = { text: 'Hello Shelf!' };

      await realtimeService.broadcastToShelf(shelfId, message);

      expect(realtimeService.roomService.broadcastToRoom).toHaveBeenCalledWith(
        `shelf_${shelfId}`,
        'shelf_message',
        message
      );
    });
  });
});
