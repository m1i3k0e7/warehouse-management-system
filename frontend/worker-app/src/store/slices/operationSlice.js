import { createSlice } from '@reduxjs/toolkit';

const initialState = {
  latestOperation: null,
  operationHistory: [],
  loading: false,
  error: null,
};

const operationSlice = createSlice({
  name: 'operation',
  initialState,
  reducers: {
    executeOperationStart: (state) => {
      state.loading = true;
      state.error = null;
    },
    executeOperationSuccess: (state, action) => {
      state.loading = false;
      state.latestOperation = action.payload;
      // Optionally add to history, depending on how history is managed
      // state.operationHistory.unshift(action.payload);
    },
    executeOperationFailure: (state, action) => {
      state.loading = false;
      state.error = action.payload;
    },
    addOperationToHistory: (state, action) => {
      state.operationHistory.unshift(action.payload);
      if (state.operationHistory.length > 50) { // Limit history size
        state.operationHistory.pop();
      }
    },
    clearOperationHistory: (state) => {
      state.operationHistory = [];
    },
  },
});

export const { executeOperationStart, executeOperationSuccess, executeOperationFailure, addOperationToHistory, clearOperationHistory } = operationSlice.actions;

export default operationSlice.reducer;
