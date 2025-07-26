import { useEffect, useRef, useState, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import io from 'socket.io-client';
import { updateShelfStatus, addOperation } from '../store/slices/shelfSlice';

export const useWebSocket = (serverUrl, options = {}) => {
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState(null);
  const socketRef = useRef(null);
  const dispatch = useDispatch();

  const connect = useCallback(() => {
    if (socketRef.current?.connected) return;

    socketRef.current = io(serverUrl, {
      transports: ['websocket', 'polling'],
      timeout: 5000,
      ...options
    });

    socketRef.current.on('connect', () => {
      console.log('WebSocket connected');
      setIsConnected(true);
      setConnectionError(null);
    });

    socketRef.current.on('disconnect', (reason) => {
      console.log('WebSocket disconnected:', reason);
      setIsConnected(false);
    });

    socketRef.current.on('connect_error', (error) => {
      console.error('WebSocket connection error:', error);
      setConnectionError(error.message);
      setIsConnected(false);
    });

    // 監聽庫存更新事件
    socketRef.current.on('inventory_update', (data) => {
      console.log('Received inventory update:', data);
      
      switch (data.type) {
        case 'material_placed':
        case 'material_removed':
          dispatch(updateShelfStatus({
            shelfId: data.data.shelfId,
            slotId: data.data.slotId,
            materialId: data.data.materialId,
            status: data.type === 'material_placed' ? 'occupied' : 'empty'
          }));
          break;
        default:
          console.warn('Unknown inventory update type:', data.type);
      }
    });

    // 監聽料架狀態
    socketRef.current.on('shelf_status', (status) => {
      console.log('Received shelf status:', status);
      dispatch(updateShelfStatus(status));
    });

    // 監聽操作響應
    socketRef.current.on('operation_response', (response) => {
      if (response.success) {
        dispatch(addOperation(response.data));
      } else {
        console.error('Operation failed:', response.error);
        // 這裡可以顯示錯誤通知
      }
    });

  }, [serverUrl, options, dispatch]);

  const disconnect = useCallback(() => {
    if (socketRef.current) {
      socketRef.current.disconnect();
      socketRef.current = null;
    }
  }, []);

  const joinShelfRoom = useCallback((shelfId, operatorId) => {
    if (socketRef.current?.connected) {
      socketRef.current.emit('join_shelf', { shelfId, operatorId });
    }
  }, []);

  const sendOperationRequest = useCallback((operationData) => {
    if (socketRef.current?.connected) {
      socketRef.current.emit('operation_request', operationData);
    }
  }, []);

  useEffect(() => {
    connect();

    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    isConnected,
    connectionError,
    joinShelfRoom,
    sendOperationRequest,
    reconnect: connect,
    disconnect
  };
};