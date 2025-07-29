import React, { useState, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Container, Typography, Box, TextField, Button, Paper, Alert, Grid } from '@mui/material';
import { setShelfId, fetchShelfStatusStart, fetchShelfStatusSuccess, fetchShelfStatusFailure } from '../store/slices/shelfSlice';
import { inventoryApi } from '../services/api';
import ShelfGrid from '../components/shelf/ShelfGrid';
import { logger } from '../utils/logger';

function ShelfManagement() {
  const dispatch = useDispatch();
  const { currentShelfId, shelfStatus, loading, error } = useSelector((state) => state.shelf);
  const [inputShelfId, setInputShelfId] = useState('shelf-1');

  useEffect(() => {
    if (currentShelfId) {
      const fetchStatus = async () => {
        dispatch(fetchShelfStatusStart());
        try {
          const response = await inventoryApi.getShelfStatus(currentShelfId);
          dispatch(fetchShelfStatusSuccess(response.data));
        } catch (err) {
          dispatch(fetchShelfStatusFailure(err.message));
          logger.error('Failed to fetch shelf status:', err);
        }
      };
      fetchStatus();
    }
  }, [currentShelfId, dispatch]);

  const handleSetShelf = () => {
    dispatch(setShelfId(inputShelfId));
  };

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Shelf Management
      </Typography>

      <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>View Shelf Status</Typography>
        <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
          <TextField
            label="Enter Shelf ID"
            variant="outlined"
            value={inputShelfId}
            onChange={(e) => setInputShelfId(e.target.value)}
            fullWidth
          />
          <Button variant="contained" onClick={handleSetShelf}>
            View Shelf
          </Button>
        </Box>

        {loading && <Typography>Loading shelf status...</Typography>}
        {error && <Alert severity="error">Error: {error}</Alert>}

        {shelfStatus && (
          <Box sx={{ mt: 3 }}>
            <Typography variant="h5" gutterBottom>Current Shelf: {shelfStatus.shelfId}</Typography>
            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <Paper elevation={1} sx={{ p: 2 }}>
                  <Typography>Total Slots: {shelfStatus.totalSlots}</Typography>
                  <Typography>Empty Slots: {shelfStatus.emptySlots}</Typography>
                  <Typography>Occupied Slots: {shelfStatus.occupiedSlots}</Typography>
                </Paper>
              </Grid>
              <Grid item xs={12}>
                <ShelfGrid />
              </Grid>
            </Grid>
          </Box>
        )}
      </Paper>
    </Container>
  );
}

export default ShelfManagement;
