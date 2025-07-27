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
  rateLimit: {
    requests: parseInt(process.env.RATE_LIMIT_REQUESTS || '100', 10), // allowed requests per window
    window: parseInt(process.env.RATE_LIMIT_WINDOW || '60', 10),     // limit window in seconds
  },
};

module.exports = config;
