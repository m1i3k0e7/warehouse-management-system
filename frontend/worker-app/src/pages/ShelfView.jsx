import React, { useEffect, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { Container, Typography, Box, CircularProgress, Alert, Grid, Paper, Button } from '@mui/material';
import ShelfGrid from '../components/shelf/ShelfGrid';
import { setShelfId, fetchShelfStatusStart, fetchShelfStatusSuccess, fetchShelfStatusFailure } from '../store/slices/shelfSlice';
import { inventoryApi } from '../services/api';
import useWebSocket from '../hooks/useWebSocket';
import { logger } from '../utils/logger';

function ShelfView() {
  const { shelfId } = useParams();
  const dispatch = useDispatch();
  const { currentShelfId, shelfStatus, loading, error } = useSelector((state) => state.shelf);
  const { joinShelfRoom } = useWebSocket();
  const [operatorId, setOperatorId] = useState('operator-1'); // This should come from user auth

  const fetchShelfStatus = useCallback(async (shelfIdToFetch) => {
    if (!shelfIdToFetch) return;

    dispatch(fetchShelfStatusStart());
    try {
      const response = await inventoryApi.getShelfStatus(shelfIdToFetch);
      dispatch(fetchShelfStatusSuccess(response.data));
    } catch (err) {
      dispatch(fetchShelfStatusFailure(err.message));
      logger.error('Failed to fetch shelf status:', err);
    }
  }, [dispatch]);

  useEffect(() => {
    if (shelfId && shelfId !== currentShelfId) {
      dispatch(setShelfId(shelfId));
      fetchShelfStatus(shelfId); // Fetch status when shelfId changes
    }
  }, [shelfId, currentShelfId, dispatch, fetchShelfStatus]);

  useEffect(() => {
    if (currentShelfId) {
      joinShelfRoom(currentShelfId, operatorId);
    }
  }, [currentShelfId, operatorId, joinShelfRoom]);

  const handleRetryFetch = () => {
    if (currentShelfId) {
      fetchShelfStatus(currentShelfId);
    }
  };

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Shelf View: {shelfId}
      </Typography>

      {loading && (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
          <CircularProgress />
          <Typography variant="subtitle1" sx={{ ml: 2 }}>Loading shelf status...</Typography>
        </Box>
      )}
      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          Error: {error}
          <Button onClick={handleRetryFetch} sx={{ ml: 2 }} variant="outlined" size="small">
            Retry
          </Button>
        </Alert>
      )}

      {shelfStatus && (
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <Paper elevation={3} sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>Shelf Summary</Typography>
              <Typography>Total Slots: {shelfStatus.totalSlots}</Typography>
              <Typography>Empty Slots: {shelfStatus.emptySlots}</Typography>
              <Typography>Occupied Slots: {shelfStatus.occupiedSlots}</Typography>
            </Paper>
          </Grid>
          <Grid item xs={12}>
            <Paper elevation={3} sx={{ p: 2 }}>
              <ShelfGrid />
            </Paper>
          </Grid>
        </Grid>
      )}
    </Container>
  );
}

export default ShelfView;
