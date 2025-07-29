import React, { useState, useEffect, useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Container, Typography, Box, TextField, Button, Paper, Alert, MenuItem, Select, InputLabel, FormControl, Grid, CircularProgress, List, ListItem, ListItemText } from '@mui/material';
import useWebSocket from '../hooks/useWebSocket';
import { executeOperationStart, executeOperationSuccess, executeOperationFailure, addOperationToHistory } from '../store/slices/operationSlice';
import { setShelfId, fetchShelfStatusStart, fetchShelfStatusSuccess, fetchShelfStatusFailure } from '../store/slices/shelfSlice';
import { inventoryApi } from '../services/api';
import ShelfGrid from '../components/shelf/ShelfGrid';
import { logger } from '../utils/logger';
import { OPERATION_TYPE } from '../utils/constants';

function MaterialOperations() {
  const dispatch = useDispatch();
  const { loading: operationLoading, error: operationError, latestOperation, operationHistory } = useSelector((state) => state.operation);
  const { currentShelfId, shelfStatus, loading: shelfLoading, error: shelfError } = useSelector((state) => state.shelf);
  const { sendOperationRequest, joinShelfRoom } = useWebSocket();

  const [operationType, setOperationType] = useState(OPERATION_TYPE.PLACEMENT);
  const [materialBarcode, setMaterialBarcode] = useState('');
  const [displayShelfId, setDisplayShelfId] = useState(currentShelfId || 'shelf-1'); // Default shelf for display
  const [slotId, setSlotId] = useState('');
  const [fromSlotId, setFromSlotId] = useState('');
  const [toSlotId, setToSlotId] = useState('');
  const [operatorId, setOperatorId] = useState('operator-1'); // This should come from user auth
  const [reason, setReason] = useState('');

  // Validation error states
  const [materialBarcodeError, setMaterialBarcodeError] = useState('');
  const [slotIdError, setSlotIdError] = useState('');
  const [fromSlotIdError, setFromSlotIdError] = useState('');
  const [toSlotIdError, setToSlotIdError] = useState('');
  const [reasonError, setReasonError] = useState('');

  const resetErrors = () => {
    setMaterialBarcodeError('');
    setSlotIdError('');
    setFromSlotIdError('');
    setToSlotIdError('');
    setReasonError('');
  };

  const validateForm = () => {
    resetErrors();
    let isValid = true;

    switch (operationType) {
      case OPERATION_TYPE.PLACEMENT:
        if (!materialBarcode) {
          setMaterialBarcodeError('Material Barcode is required');
          isValid = false;
        }
        if (!slotId) {
          setSlotIdError('Target Slot ID is required');
          isValid = false;
        }
        break;
      case OPERATION_TYPE.REMOVAL:
        if (!slotId) {
          setSlotIdError('Slot ID to Remove From is required');
          isValid = false;
        }
        // Reason is optional for removal
        break;
      case OPERATION_TYPE.MOVE:
        if (!fromSlotId) {
          setFromSlotIdError('From Slot ID is required');
          isValid = false;
        }
        if (!toSlotId) {
          setToSlotIdError('To Slot ID is required');
          isValid = false;
        }
        // Reason is optional for move
        break;
      default:
        isValid = false; // Should not happen with controlled select
    }
    return isValid;
  };

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

  // useEffect(() => {
  //   // Join WebSocket room for the current shelf being displayed
  //   if (displayShelfId) {
  //     joinShelfRoom(displayShelfId, operatorId);
  //     fetchShelfStatus(displayShelfId);
  //   }
  // }, [displayShelfId, operatorId, joinShelfRoom, fetchShelfStatus]);

  const handleOperation = async () => {
    if (!validateForm()) {
      return; // Stop if validation fails
    }

    dispatch(executeOperationStart());
    let payload = { operatorId };
    let targetShelfId = '';

    switch (operationType) {
      case OPERATION_TYPE.PLACEMENT:
        payload = { ...payload, type: operationType, materialBarcode, slotId };
        targetShelfId = slotId.split('-')[0]; // Assuming slotId format like shelfId-row-col
        break;
      case OPERATION_TYPE.REMOVAL:
        payload = { ...payload, type: operationType, slotId, reason };
        targetShelfId = slotId.split('-')[0];
        break;
      case OPERATION_TYPE.MOVE:
        payload = { ...payload, type: operationType, fromSlotId, toSlotId, reason };
        targetShelfId = fromSlotId.split('-')[0]; // Or toSlotId.split('-')[0] if cross-shelf move
        break;
      default:
        dispatch(executeOperationFailure('Unknown operation type'));
        return;
    }

    try {
      /*  sendOperationRequest(payload) 
          -> dispatch({ type: 'websocket/operationRequest', payload: { ...operationData, requestId } });
          -> socket.emit('operation_request', payload);
          -> realtimeService.handleOperationRequest(socket, payload);
          -> realtimeService.processOperation(payload);
          -> inventoryAPIService.placeMaterial(payload) / removeMaterial(payload) / moveMaterial(payload)
          -> axios.post(`${this.inventoryServiceUrl}/materials/place`, data);
          -> inventoryService.MaterialHandler.placeMaterial(data) -> inventoryService.MaterialHandler.placeMaterialHandler.Handle(context, command);
          -> invertoryService.placeMaterial()
          -> inventoryService.executePlaceMaterial() -> begin transaction -> insert material into slot -> commit transaction
          -> inventoryService.auditService.LogSuccessfulOperation() -> auditService.publishAuditLog()
          -> inventoryService.eventService.publishEvent() -> send Kafka message
          -> realtimeService.kafkaController.handleMessage()
          -> realtimeService.inventoryEventHandler.handle() -> realtimeService.broadcastInventoryUpdate()
          -> roomService.broadcastToRoom() -> socket.to(roomId).emit('system_event', payload);
          -> roomService.updateRealtimeStats() -> update Redis cache
          -> frontend.websocketMiddleware.socket.on('system_event')
          -> dispatch(fetchShelfStatusSuccess(data)) -> update shelf status in Redux store
          -> dispatch(addOperationToHistory({ ...payload, timestamp: new Date().toISOString(), status: 'success' }));
      */
      const response = await sendOperationRequest(payload);
      dispatch(executeOperationSuccess(response)); // Use response from backend if available
      dispatch(addOperationToHistory({ ...payload, timestamp: new Date().toISOString(), status: 'success' }));
      
      // After successful operation, update the displayed shelf status
      if (targetShelfId) {
        setDisplayShelfId(targetShelfId);
        fetchShelfStatus(targetShelfId);
      }

      // Clear form fields and errors
      setMaterialBarcode('');
      setSlotId('');
      setFromSlotId('');
      setToSlotId('');
      setReason('');
      resetErrors();
    } catch (err) {
      dispatch(executeOperationFailure(err.message));
      logger.error('Operation failed:', err);
      dispatch(addOperationToHistory({ ...payload, timestamp: new Date().toISOString(), status: 'failed', error: err.message }));
    }
  };

  const handleRetryFetchShelf = () => {
    if (displayShelfId) {
      fetchShelfStatus(displayShelfId);
    }
  };

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Material Operations
      </Typography>

      <Grid container spacing={3}>
        {/* Left Panel: Operation Forms and History */}
        <Grid item xs={12} md={6}>
          <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
            <Typography variant="h6" gutterBottom>Execute New Operation</Typography>
            <FormControl fullWidth sx={{ mb: 2 }}>
              <InputLabel id="operation-type-label">Operation Type</InputLabel>
              <Select
                labelId="operation-type-label"
                value={operationType}
                label="Operation Type"
                onChange={(e) => setOperationType(e.target.value)}
              >
                {Object.values(OPERATION_TYPE).map((type) => (
                  <MenuItem key={type} value={type}>
                    {type.replace(/_/g, ' ').toUpperCase()}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {operationType === OPERATION_TYPE.PLACEMENT && (
              <Box sx={{ mb: 2 }}>
                <TextField
                  fullWidth
                  label="Material Barcode"
                  value={materialBarcode}
                  onChange={(e) => {
                    setMaterialBarcode(e.target.value);
                    setMaterialBarcodeError(''); // Clear error on change
                  }}
                  sx={{ mb: 2 }}
                  error={!!materialBarcodeError}
                  helperText={materialBarcodeError}
                />
                <TextField
                  fullWidth
                  label="Target Slot ID"
                  value={slotId}
                  onChange={(e) => {
                    setSlotId(e.target.value);
                    setSlotIdError(''); // Clear error on change
                  }}
                  error={!!slotIdError}
                  helperText={slotIdError}
                />
              </Box>
            )}

            {operationType === OPERATION_TYPE.REMOVAL && (
              <Box sx={{ mb: 2 }}>
                <TextField
                  fullWidth
                  label="Slot ID to Remove From"
                  value={slotId}
                  onChange={(e) => {
                    setSlotId(e.target.value);
                    setSlotIdError(''); // Clear error on change
                  }}
                  sx={{ mb: 2 }}
                  error={!!slotIdError}
                  helperText={slotIdError}
                />
                <TextField
                  fullWidth
                  label="Reason for Removal"
                  value={reason}
                  onChange={(e) => {
                    setReason(e.target.value);
                    setReasonError(''); // Clear error on change
                  }}
                  error={!!reasonError}
                  helperText={reasonError}
                />
              </Box>
            )}

            {operationType === OPERATION_TYPE.MOVE && (
              <Box sx={{ mb: 2 }}>
                <TextField
                  fullWidth
                  label="From Slot ID"
                  value={fromSlotId}
                  onChange={(e) => {
                    setFromSlotId(e.target.value);
                    setFromSlotIdError(''); // Clear error on change
                  }}
                  sx={{ mb: 2 }}
                  error={!!fromSlotIdError}
                  helperText={fromSlotIdError}
                />
                <TextField
                  fullWidth
                  label="To Slot ID"
                  value={toSlotId}
                  onChange={(e) => {
                    setToSlotId(e.target.value);
                    setToSlotIdError(''); // Clear error on change
                  }}
                  sx={{ mb: 2 }}
                  error={!!toSlotIdError}
                  helperText={toSlotIdError}
                />
                <TextField
                  fullWidth
                  label="Reason for Move"
                  value={reason}
                  onChange={(e) => {
                    setReason(e.target.value);
                    setReasonError(''); // Clear error on change
                  }}
                  error={!!reasonError}
                  helperText={reasonError}
                />
              </Box>
            )}

            <Button variant="contained" onClick={handleOperation} disabled={operationLoading} fullWidth>
              {operationLoading ? 'Processing...' : 'Execute Operation'}
            </Button>

            {operationError && (
              <Alert severity="error" sx={{ mt: 2 }}>
                Error: {operationError}
              </Alert>
            )}
            {latestOperation && !operationLoading && !operationError && (
              <Alert severity="success" sx={{ mt: 2 }}>
                Operation request sent successfully!
                {latestOperation.type === OPERATION_TYPE.PLACEMENT && ` Material ${latestOperation.materialBarcode} to ${latestOperation.slotId}`}
                {latestOperation.type === OPERATION_TYPE.REMOVAL && ` Material from ${latestOperation.slotId}`}
                {latestOperation.type === OPERATION_TYPE.MOVE && ` Material from ${latestOperation.fromSlotId} to ${latestOperation.toSlotId}`}
              </Alert>
            )}
          </Paper>

          <Paper elevation={3} sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>Operation History</Typography>
            {operationHistory.length === 0 ? (
              <Typography>No operations recorded yet.</Typography>
            ) : (
              <List>
                {operationHistory.map((op, index) => (
                  <ListItem key={index} divider>
                    <ListItemText
                      primary={
                        <React.Fragment>
                          <Typography component="span" variant="body2" color="text.primary">
                            <strong>Type:</strong> {op.type.replace(/_/g, ' ').toUpperCase()}
                          </Typography>
                          {op.materialBarcode && <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}><strong>Material:</strong> {op.materialBarcode}</Typography>}
                        </React.Fragment>
                      }
                      secondary={
                        <React.Fragment>
                          {op.slotId && <Typography component="span" variant="body2" color="text.secondary"><strong>Slot:</strong> {op.slotId}</Typography>}
                          {op.fromSlotId && <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}><strong>From:</strong> {op.fromSlotId}</Typography>}
                          {op.toSlotId && <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}><strong>To:</strong> {op.toSlotId}</Typography>}
                          <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}><strong>Operator:</strong> {op.operatorId}</Typography>
                          <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}><strong>Time:</strong> {new Date(op.timestamp).toLocaleString()}</Typography>
                        </React.Fragment>
                      }
                    />
                  </ListItem>
                ))}
              </List>
            )}
          </Paper>
        </Grid>

        {/* Right Panel: Displayed Shelf Information */}
        <Grid item xs={12} md={6}>
          <Paper elevation={3} sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>Affected Shelf Status</Typography>
            {shelfLoading && (
              <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
                <CircularProgress />
                <Typography variant="subtitle1" sx={{ ml: 2 }}>Loading shelf status...</Typography>
              </Box>
            )}
            {shelfError && (
              <Alert severity="error" sx={{ mt: 2 }}>
                Error: {shelfError}
                <Button onClick={handleRetryFetchShelf} sx={{ ml: 2 }} variant="outlined" size="small">
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
              </Box>
            ) : (
              <Typography variant="body1" color="text.secondary">No shelf affected by the last operation, or not loaded.</Typography>
            )}
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
}

export default MaterialOperations;
