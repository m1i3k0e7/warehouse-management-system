import { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import socket from '../services/websocket';
import { logger } from '../utils/logger';

const useWebSocket = () => {
  const dispatch = useDispatch();

  useEffect(() => {
    // Dispatch an action to connect WebSocket when component mounts
    dispatch({ type: 'websocket/connect' });

    // Clean up WebSocket connection when component unmounts
    return () => {
      dispatch({ type: 'websocket/disconnect' });
    };
  }, [dispatch]);

  const joinShelfRoom = (shelfId, operatorId) => {
    dispatch({ type: 'websocket/joinShelf', payload: { shelfId, operatorId } });
  };

  const sendOperationRequest = (operationData) => {
    dispatch({ type: 'websocket/operationRequest', payload: operationData });
  };

  return { joinShelfRoom, sendOperationRequest };
};

export default useWebSocket;
