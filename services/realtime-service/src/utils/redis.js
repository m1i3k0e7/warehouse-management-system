const { createClient } = require('redis');
const config = require('../config');
const logger = require('./logger');

const client = createClient({
  url: config.redis.url
});

client.on('connect', () => logger.info('Redis client connected'));
client.on('ready', () => logger.info('Redis client ready'));
client.on('end', () => logger.warn('Redis client disconnected'));
client.on('reconnecting', () => logger.info('Redis client reconnecting...'));
client.on('error', (err) => logger.error('Redis client error:', err));

(async () => {
  await client.connect();
})();

module.exports = client;
