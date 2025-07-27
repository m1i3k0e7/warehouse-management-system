const redis = require('../utils/redis');
const logger = require('../utils/logger');
const config = require('../config');

// limit the number of requests per IP or user within a certain time frame
const rateLimitMiddleware = (socket, next) => {
  const ip = socket.handshake.address; // get the IP address of the client
  const limit = config.rateLimit.requests; // allowed requests per window
  const window = config.rateLimit.window; // lmaximum time window in seconds
  const key = `ratelimit:${ip}`;

  redis.incr(key)
    .then(count => {
      if (count === 1) {
        // set the expiration time for the key if it's the first request
        redis.expire(key, window);
      }

      if (count > limit) {
        logger.warn(`Rate limit exceeded for IP: ${ip}`);
        return next(new Error('Rate limit exceeded'));
      }

      next();
    })
    .catch(err => {
      logger.error(`Rate limit middleware error for IP ${ip}:`, err);
      next(new Error('Internal server error during rate limiting'));
    });
};

module.exports = rateLimitMiddleware;
