const logger = require('../utils/logger');

const authMiddleware = (socket, next) => {
  const token = socket.handshake.auth.token;

  if (!token) {
    logger.warn(`Authentication failed: No token provided for socket ${socket.id}`);
    return next(new Error('Authentication error: No token provided'));
  }

  // assuming a simple token-based authentication
  if (token === 'your-secret-auth-token') {
    // parse the token and attach user info to the socket
    socket.user = { id: 'some-user-id', role: 'worker' }; // example user object
    logger.info(`Client authenticated: ${socket.id}, User: ${socket.user.id}`);
    next();
  } else {
    logger.warn(`Authentication failed: Invalid token for socket ${socket.id}`);
    next(new Error('Authentication error: Invalid token'));
  }
};

module.exports = authMiddleware;
