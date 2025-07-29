import { createSlice } from '@reduxjs/toolkit';

const initialState = {
  currentShelfId: null,
  shelfStatus: null, // { shelfId, totalSlots, emptySlots, occupiedSlots, slots: [] }
  loading: false,
  error: null,
};

const shelfSlice = createSlice({
  name: 'shelf',
  initialState,
  reducers: {
    setShelfId: (state, action) => {
      state.currentShelfId = action.payload;
    },
    fetchShelfStatusStart: (state) => {
      state.loading = true;
      state.error = null;
    },
    fetchShelfStatusSuccess: (state, action) => {
      state.loading = false;
      state.shelfStatus = action.payload;
    },
    fetchShelfStatusFailure: (state, action) => {
      state.loading = false;
      state.error = action.payload;
    },
    updateSlotStatus: (state, action) => {
      const { slotId, newStatus, materialId } = action.payload;
      if (state.shelfStatus && state.shelfStatus.slots) {
        const slotIndex = state.shelfStatus.slots.findIndex(slot => slot.ID === slotId);
        if (slotIndex !== -1) {
          const oldStatus = state.shelfStatus.slots[slotIndex].Status;
          state.shelfStatus.slots[slotIndex].Status = newStatus;
          state.shelfStatus.slots[slotIndex].MaterialID = materialId || null;

          // Update counts
          if (oldStatus === 'empty' && newStatus === 'occupied') {
            state.shelfStatus.emptySlots--;
            state.shelfStatus.occupiedSlots++;
          } else if (oldStatus === 'occupied' && newStatus === 'empty') {
            state.shelfStatus.emptySlots++;
            state.shelfStatus.occupiedSlots--;
          }
        }
      }
    },
  },
});

export const { setShelfId, fetchShelfStatusStart, fetchShelfStatusSuccess, fetchShelfStatusFailure, updateSlotStatus } = shelfSlice.actions;

export default shelfSlice.reducer;
