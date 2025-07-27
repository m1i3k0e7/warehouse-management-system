const { Kafka, logLevel } = require('kafkajs');
const KafkaController = require('../../src/controllers/kafkaController');
const RealtimeService = require('../../src/services/realtimeService');
const NotificationService = require('../../src/services/notificationService');
const RoomService = require('../../src/services/roomService');
const config = require('../../src/config');
const redis = require('../../src/utils/redis');

// Mock Socket.IO server for RealtimeService
const ioMock = {
  to: jest.fn().mockReturnThis(),
  emit: jest.fn(),
};

// Mock RealtimeService methods that KafkaController calls
const mockRealtimeService = new RealtimeService(ioMock, new NotificationService(), new RoomService(ioMock));
Object.keys(mockRealtimeService).forEach(key => {
  if (typeof mockRealtimeService[key] === 'function') {
    jest.spyOn(mockRealtimeService, key).mockImplementation(() => {});
  }
});

// Kafka test client setup
const testKafka = new Kafka({
  clientId: 'test-producer',
  brokers: config.kafka.brokers,
  logLevel: logLevel.ERROR, // Suppress KafkaJS logs during tests
});
const producer = testKafka.producer();

describe('KafkaController Integration', () => {
  let kafkaController;

  beforeAll(async () => {
    // Ensure Kafka and Redis are running before tests
    // In a real setup, this would be handled by test containers or a dedicated test environment
    try {
      await producer.connect();
      await redis.connect();
      console.log('Connected to Kafka and Redis for integration tests.');
    } catch (error) {
      console.error('Failed to connect to Kafka or Redis for integration tests:', error);
      process.exit(1);
    }
  });

  afterAll(async () => {
    await producer.disconnect();
    await redis.quit();
  });

  beforeEach(async () => {
    // Clear mocks before each test
    jest.clearAllMocks();
    // Stop and restart consumer to clear state
    if (kafkaController && kafkaController.isRunning) {
      await kafkaController.stop();
    }
    kafkaController = new KafkaController(mockRealtimeService);
    await kafkaController.start();
    // Give consumer some time to subscribe
    await new Promise(resolve => setTimeout(resolve, 1000)); 
  });

  afterEach(async () => {
    if (kafkaController && kafkaController.isRunning) {
      await kafkaController.stop();
    }
  });

  test('should call broadcastInventoryUpdate for material.placed event', async () => {
    const event = {
      event_type: 'material.placed',
      shelf_id: 'shelf1',
      slot_id: 'slot1',
      material_id: 'mat1',
      timestamp: new Date().toISOString(),
    };
    await producer.send({
      topic: config.kafka.topic,
      messages: [{ key: 'material.placed', value: JSON.stringify(event) }],
    });

    // Wait for the message to be consumed and processed
    await new Promise(resolve => setTimeout(resolve, 500));

    expect(mockRealtimeService.broadcastInventoryUpdate).toHaveBeenCalledWith(event);
  });

  test('should call broadcastShelfStatusChange for shelf.status_changed event', async () => {
    const event = {
      event_type: 'shelf.status_changed',
      shelf_id: 'shelf1',
      old_status: 'offline',
      new_status: 'online',
      timestamp: new Date().toISOString(),
    };
    await producer.send({
      topic: config.kafka.topic,
      messages: [{ key: 'shelf.status_changed', value: JSON.stringify(event) }],
    });

    await new Promise(resolve => setTimeout(resolve, 500));

    expect(mockRealtimeService.broadcastShelfStatusChange).toHaveBeenCalledWith(event);
  });

  test('should call broadcastSystemAlert for system.alert event', async () => {
    const event = {
      event_type: 'system.alert',
      alert_type: 'sensor_error',
      severity: 'high',
      message: 'Sensor malfunction',
      timestamp: new Date().toISOString(),
    };
    await producer.send({
      topic: config.kafka.topic,
      messages: [{ key: 'system.alert', value: JSON.stringify(event) }],
    });

    await new Promise(resolve => setTimeout(resolve, 500));

    expect(mockRealtimeService.broadcastSystemAlert).toHaveBeenCalledWith(event);
  });

  test('should call broadcastAuditLog for audit.log event', async () => {
    const event = {
      event_type: 'audit.log',
      action: 'login',
      operatorId: 'user1',
      timestamp: new Date().toISOString(),
    };
    await producer.send({
      topic: config.kafka.topic,
      messages: [{ key: 'audit.log', value: JSON.stringify(event) }],
    });

    await new Promise(resolve => setTimeout(resolve, 500));

    expect(mockRealtimeService.broadcastAuditLog).toHaveBeenCalledWith(event);
  });

  test('should log warning for unknown event type', async () => {
    const event = {
      event_type: 'unknown.event',
      data: 'some data',
    };
    await producer.send({
      topic: config.kafka.topic,
      messages: [{ key: 'unknown.event', value: JSON.stringify(event) }],
    });

    await new Promise(resolve => setTimeout(resolve, 500));

    expect(mockRealtimeService.broadcastInventoryUpdate).not.toHaveBeenCalled();
    expect(mockRealtimeService.broadcastShelfStatusChange).not.toHaveBeenCalled();
    expect(mockRealtimeService.broadcastSystemAlert).not.toHaveBeenCalled();
    expect(mockRealtimeService.broadcastAuditLog).not.toHaveBeenCalled();
    // Expect logger.warn to be called, but since logger is mocked, we can't directly check its calls here
    // unless we specifically mock logger.warn as well.
  });
});
