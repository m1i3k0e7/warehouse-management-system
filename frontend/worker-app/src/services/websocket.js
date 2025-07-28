import { io } from 'socket.io-client';
import { WEBSOCKET_URL, AUTH_TOKEN } from '../utils/constants';
import { logger } from '../utils/logger';

const socket = io(WEBSOCKET_URL, {
  auth: {
    token: AUTH_TOKEN,
  },
  transports: ['websocket', 'polling'],
});

socket.on('connect', () => {
  logger.info('WebSocket connected', socket.id);
});

socket.on('disconnect', (reason) => {
  logger.warn('WebSocket disconnected:', reason);
});

socket.on('connect_error', (error) => {
  logger.error('WebSocket connection error:', error.message);
});

export default socket;
