import { configureStore } from '@reduxjs/toolkit';
import dashboardReducer from './slices/dashboardSlice';
import shelfReducer from './slices/shelfSlice';
import operationReducer from './slices/operationSlice';
import notificationReducer from './slices/notificationSlice';
import websocketMiddleware from './middleware/websocketMiddleware';

const store = configureStore({
  reducer: {
    dashboard: dashboardReducer,
    shelf: shelfReducer,
    operation: operationReducer,
    notification: notificationReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(websocketMiddleware),
});

export default store;
