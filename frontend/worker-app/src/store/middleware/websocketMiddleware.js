import socket from '../../services/websocket';
import { updateSlotStatus } from '../slices/shelfSlice';
import { addOperationToHistory } from '../slices/operationSlice';
import { logger } from '../../utils/logger';

const websocketMiddleware = (store) => (next) => (action) => {
  switch (action.type) {
    case 'websocket/connect':
      socket.connect();
      break;
    case 'websocket/disconnect':
      socket.disconnect();
      break;
    case 'websocket/joinShelf':
      socket.emit('join_shelf', action.payload);
      break;
    case 'websocket/operationRequest':
      socket.emit('operation_request', action.payload);
      break;
    default:
      break;
  }

  // Handle incoming WebSocket messages
  socket.on('inventory_update', (payload) => {
    logger.info('Received inventory_update:', payload);
    const { shelfId, slotId, materialId, type } = payload.data;
    let newStatus;
    if (type === 'material_placed') {
      newStatus = 'occupied';
    } else if (type === 'material_removed') {
      newStatus = 'empty';
    }
    store.dispatch(updateSlotStatus({ slotId, newStatus, materialId }));
    store.dispatch(addOperationToHistory({ ...payload.data, type }));
  });

  socket.on('shelf_status', (payload) => {
    logger.info('Received shelf_status:', payload);
    // Dispatch action to update shelf status in store
    // store.dispatch(updateShelfStatus(payload));
  });

  socket.on('operation_response', (payload) => {
    logger.info('Received operation_response:', payload);
    // Handle operation response, e.g., show success/failure message
    // store.dispatch(handleOperationResponse(payload));
  });

  socket.on('connect_error', (error) => {
    logger.error('WebSocket connect_error:', error);
  });

  socket.on('error', (error) => {
    logger.error('WebSocket error:', error);
  });

  return next(action);
};

export default websocketMiddleware;
