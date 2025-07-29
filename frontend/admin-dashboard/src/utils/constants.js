export const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080/api/v1';
export const WEBSOCKET_URL = process.env.REACT_APP_WEBSOCKET_URL || 'http://localhost:3001';
export const AUTH_TOKEN = process.env.REACT_APP_AUTH_TOKEN || 'your-secret-auth-token';

export const SHELF_STATUS = {
  EMPTY: 'empty',
  OCCUPIED: 'occupied',
  RESERVED: 'reserved',
  MAINTENANCE: 'maintenance',
};

export const MATERIAL_STATUS = {
  AVAILABLE: 'available',
  IN_USE: 'in_use',
  RESERVED: 'reserved',
  MAINTENANCE: 'maintenance',
};

export const OPERATION_TYPE = {
  PLACEMENT: 'placement',
  REMOVAL: 'removal',
  MOVE: 'move',
  RESERVATION: 'reservation',
};
