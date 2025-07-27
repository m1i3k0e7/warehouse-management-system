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
    retryInitialTime: parseInt(process.env.KAFKA_RETRY_INITIAL_TIME || '100', 10),
    retryRetries: parseInt(process.env.KAFKA_RETRY_RETRIES || '8', 10),
  },
  redis: {
    url: process.env.REDIS_URL || 'redis://localhost:6379',
    sessionExpiration: parseInt(process.env.REDIS_SESSION_EXPIRATION || '3600', 10), // seconds
    shelfStatusCacheExpiration: parseInt(process.env.REDIS_SHELF_STATUS_CACHE_EXPIRATION || '600', 10), // seconds
    realtimeStatsExpiration: parseInt(process.env.REDIS_REALTIME_STATS_EXPIRATION || '604800', 10), // seconds (7 days)
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
  auth: {
    secretToken: process.env.AUTH_SECRET_TOKEN || 'super-secret-jwt-key',
  },
};

module.exports = config;
