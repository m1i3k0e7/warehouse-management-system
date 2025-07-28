import React, { useState, useEffect, useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Container, Typography, Box, TextField, Button, Paper, Alert, Grid, List, ListItem, ListItemText } from '@mui/material';
import useWebSocket from '../hooks/useWebSocket';
import { setShelfId, fetchShelfStatusStart, fetchShelfStatusSuccess, fetchShelfStatusFailure } from '../store/slices/shelfSlice';
import { inventoryApi } from '../services/api';
import ShelfGrid from '../components/shelf/ShelfGrid';
import { logger } from '../utils/logger';
import { Link, useNavigate } from 'react-router-dom';

function Dashboard() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { currentShelfId, shelfStatus, loading, error } = useSelector((state) => state.shelf);
  const { joinShelfRoom } = useWebSocket();

  const [inputShelfId, setInputShelfId] = useState('shelf-1'); // Default for testing
  const [operatorId, setOperatorId] = useState('operator-1'); // Default for testing

  // Example of quick access shelves (can be fetched from API or local storage)
  const quickAccessShelves = [
    { id: 'shelf-1', name: 'Main Storage A' },
    { id: 'shelf-2', name: 'Main Storage B' },
    { id: 'shelf-3', name: 'Receiving Area' },
  ];

  const fetchShelfStatus = async (shelfIdToFetch) => {
    if (!shelfIdToFetch) return;

    dispatch(fetchShelfStatusStart());
    try {
      const response = await inventoryApi.getShelfStatus(shelfIdToFetch);
      dispatch(fetchShelfStatusSuccess(response.data));
    } catch (err) {
      dispatch(fetchShelfStatusFailure(err.message));
      logger.error('Failed to fetch shelf status:', err);
    }
  };

  useEffect(() => {
    // Only connect WebSocket when currentShelfId is set
    if (currentShelfId) {
      joinShelfRoom(currentShelfId, operatorId);
    }
  }, [currentShelfId, operatorId, joinShelfRoom]);

  const handleSetShelf = () => {
    if (inputShelfId) {
      dispatch(setShelfId(inputShelfId));
      fetchShelfStatus(inputShelfId);
    }
  };

  const handleQuickAccessClick = (shelfId) => {
    dispatch(setShelfId(shelfId));
    fetchShelfStatus(shelfId);
  };

  const handleRetryFetch = () => {
    if (currentShelfId) {
      fetchShelfStatus(currentShelfId);
    }
  };

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Warehouse Operations Dashboard
      </Typography>

      <Grid container spacing={3}>
        {/* Left Panel: Search and Quick Access */}
        <Grid item xs={12} md={4}>
          <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
            <Typography variant="h6" gutterBottom>Search Shelf</Typography>
            <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
              <TextField
                label="Enter Shelf ID"
                variant="outlined"
                value={inputShelfId}
                onChange={(e) => setInputShelfId(e.target.value)}
                fullWidth
              />
              <Button variant="contained" onClick={handleSetShelf}>
                Search
              </Button>
            </Box>

            <Typography variant="h6" gutterBottom>Quick Access</Typography>
            <List>
              {quickAccessShelves.map((shelf) => (
                <ListItem button key={shelf.id} onClick={() => handleQuickAccessClick(shelf.id)}>
                  <ListItemText primary={shelf.name} secondary={`ID: ${shelf.id}`} />
                </ListItem>
              ))}
            </List>
          </Paper>

          <Paper elevation={3} sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>Material Operations</Typography>
            <Typography variant="body1" sx={{ mb: 2 }}>
              Perform material placement, removal, or movement.
            </Typography>
            <Button variant="contained" component={Link} to="/operations" fullWidth>
              Go to Operations Page
            </Button>
          </Paper>
        </Grid>

        {/* Right Panel: Displayed Shelf Information */}
        <Grid item xs={12} md={8}>
          <Paper elevation={3} sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>Displayed Shelf Information</Typography>
            {loading && (
              <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
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

            {shelfStatus ? (
              <Box>
                <Typography variant="h5" gutterBottom>Current Shelf: {shelfStatus.shelfId}</Typography>
                <Grid container spacing={2} sx={{ mb: 2 }}>
                  <Grid item xs={12} sm={6}>
                    <Paper elevation={1} sx={{ p: 2 }}>
                      <Typography>Total Slots: {shelfStatus.totalSlots}</Typography>
                      <Typography>Empty Slots: {shelfStatus.emptySlots}</Typography>
                      <Typography>Occupied Slots: {shelfStatus.occupiedSlots}</Typography>
                    </Paper>
                  </Grid>
                </Grid>
                <ShelfGrid />
                <Box sx={{ mt: 3, textAlign: 'right' }}>
                  <Button variant="outlined" component={Link} to={`/shelf/${shelfStatus.shelfId}`}>
                    View Detailed Shelf
                  </Button>
                </Box>
              </Box>
            ) : (
              <Typography variant="body1" color="text.secondary">No shelf selected or loaded.</Typography>
            )}
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
}

export default Dashboard;

