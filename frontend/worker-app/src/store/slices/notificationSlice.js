import { createSlice } from '@reduxjs/toolkit';

const initialState = {
  messages: [],
};

const notificationSlice = createSlice({
  name: 'notification',
  initialState,
  reducers: {
    addNotification: (state, action) => {
      state.messages.unshift({ id: Date.now(), text: action.payload, timestamp: new Date().toISOString() });
      if (state.messages.length > 5) { // Keep only the latest 5 messages
        state.messages.pop();
      }
    },
    clearNotifications: (state) => {
      state.messages = [];
    },
  },
});

export const { addNotification, clearNotifications } = notificationSlice.actions;

export default notificationSlice.reducer;
