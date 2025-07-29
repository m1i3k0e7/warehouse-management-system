import { useEffect, useRef, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import { v4 as uuidv4 } from 'uuid';
import { setHandleOperationResponseCallback } from '../store/middleware/websocketMiddleware';

const useWebSocket = () => {
  const dispatch = useDispatch();
  const callbacks = useRef({});

  const handleOperationResponse = useCallback((payload) => {
    const { requestId, success, data, error } = payload;
    if (callbacks.current[requestId]) {
      if (success) {
        callbacks.current[requestId].resolve(data);
      } else {
        callbacks.current[requestId].reject(new Error(error || 'Unknown error'));
      }
      delete callbacks.current[requestId];
    }
  }, []); // Empty dependency array because it only uses callbacks.current which is a ref

  useEffect(() => {
    dispatch({ type: 'websocket/connect' });
    setHandleOperationResponseCallback(handleOperationResponse);

    return () => {
      dispatch({ type: 'websocket/disconnect' });
      setHandleOperationResponseCallback(null);
    };
  }, [dispatch, handleOperationResponse]);

  const joinShelfRoom = (shelfId, operatorId) => {
    dispatch({ type: 'websocket/joinShelf', payload: { shelfId, operatorId } });
  };

  const sendOperationRequest = (operationData) => {
    return new Promise((resolve, reject) => {
      const requestId = uuidv4();
      callbacks.current[requestId] = { resolve, reject };
      dispatch({ type: 'websocket/operationRequest', payload: { ...operationData, requestId } });

      setTimeout(() => {
        if (callbacks.current[requestId]) {
          delete callbacks.current[requestId];
          reject(new Error('Operation request timed out'));
        }
      }, 10000);
    });
  };

  return { joinShelfRoom, sendOperationRequest };
};

export default useWebSocket;
