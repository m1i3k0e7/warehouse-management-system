import { createSlice } from '@reduxjs/toolkit';

const initialState = {
  stats: {
    totalShelves: 0,
    totalSlots: 0,
    occupiedSlots: 0,
    emptySlots: 0,
    materials: 0,
    operationsToday: 0,
  },
  loading: false,
  error: null,
};

const dashboardSlice = createSlice({
  name: 'dashboard',
  initialState,
  reducers: {
    fetchStatsStart: (state) => {
      state.loading = true;
      state.error = null;
    },
    fetchStatsSuccess: (state, action) => {
      state.loading = false;
      state.stats = action.payload;
    },
    fetchStatsFailure: (state, action) => {
      state.loading = false;
      state.error = action.payload;
    },
    updateStats: (state, action) => {
      // Logic to update stats based on real-time events
      // For example, increment operationsToday on a new operation event
    },
  },
});

export const { fetchStatsStart, fetchStatsSuccess, fetchStatsFailure, updateStats } = dashboardSlice.actions;

export default dashboardSlice.reducer;
