const config = {
  port: process.env.PORT || 3001,
  cors: {
    origin: process.env.CORS_ORIGIN || '*'
  },
  kafka: {
    clientId: process.env.KAFKA_CLIENT_ID || 'realtime-service',
    brokers: (process.env.KAFKA_BROKERS || 'localhost:9092').split(','),
    topic: process.env.KAFKA_TOPIC || 'warehouse.events',
    groupId: process.env.KAFKA_GROUP_ID || 'realtime-service-group',
  },
  redis: {
    url: process.env.REDIS_URL || 'redis://localhost:6379',
  },
  services: {
    inventory: {
      url: process.env.INVENTORY_SERVICE_URL || 'http://localhost:8080/api/v1',
    },
    // Add other services here as needed
  },
  logLevel: process.env.LOG_LEVEL || 'info',
};

module.exports = config;
