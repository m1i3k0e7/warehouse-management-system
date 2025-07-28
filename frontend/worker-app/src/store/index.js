import { configureStore } from '@reduxjs/toolkit';
import shelfReducer from './slices/shelfSlice';
import operationReducer from './slices/operationSlice';
import userReducer from './slices/userSlice';
import websocketMiddleware from './middleware/websocketMiddleware';

const store = configureStore({
  reducer: {
    shelf: shelfReducer,
    operation: operationReducer,
    user: userReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(websocketMiddleware),
});

export default store;
