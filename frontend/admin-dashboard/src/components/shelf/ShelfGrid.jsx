import React from 'react';
import { useSelector } from 'react-redux';
import { Box, Typography, Grid, Paper } from '@mui/material';
import { SHELF_STATUS } from '../../utils/constants';

function ShelfGrid() {
  const shelfStatus = useSelector((state) => state.shelf.shelfStatus);

  if (!shelfStatus) {
    return (
      <Box sx={{ p: 2, textAlign: 'center', color: 'text.secondary' }}>
        <Typography variant="body1">Select a shelf to view its status.</Typography>
      </Box>
    );
  }

  const { shelfId, slots } = shelfStatus;

  // Group slots by row for easier rendering
  const rows = {};
  slots.forEach(slot => {
    if (!rows[slot.Row]) {
      rows[slot.Row] = [];
    }
    rows[slot.Row].push(slot);
  });

  return (
    <Box sx={{ mt: 2 }}>
      <Typography variant="h6" gutterBottom>Shelf: {shelfId}</Typography>
      <Box sx={{ overflowX: 'auto' }}>
        <Grid container spacing={0.5} wrap="nowrap" sx={{ minWidth: 'max-content' }}>
          {/* Column Headers */}
          <Grid item>
            <Box sx={{ width: 40, height: 40, display: 'flex', alignItems: 'center', justifyContent: 'center' }}></Box>
          </Grid>
          {Array.from({ length: Math.max(...Object.values(rows).map(row => row.length)) }, (_, i) => (
            <Grid item key={`col-header-${i}`}>
              <Box sx={{ width: 40, height: 40, display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 'bold' }}>
                {i + 1}
              </Box>
            </Grid>
          ))}

          {/* Rows */}
          {Object.keys(rows).sort((a, b) => a - b).map(rowNum => (
            <Grid container item key={`row-${rowNum}`} spacing={0.5} wrap="nowrap" sx={{ minWidth: 'max-content' }}>
              <Grid item>
                <Box sx={{ width: 40, height: 40, display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 'bold' }}>
                  {rowNum}
                </Box>
              </Grid>
              {rows[rowNum].sort((a, b) => a.Column - b.Column).map(slot => (
                <Grid item key={slot.ID}>
                  <Paper
                    sx={{
                      width: 40,
                      height: 40,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      bgcolor: slot.Status === SHELF_STATUS.EMPTY ? '#e0ffe0' : 
                               slot.Status === SHELF_STATUS.OCCUPIED ? '#ffe0e0' : 
                               slot.Status === SHELF_STATUS.RESERVED ? '#e0e0ff' : 
                               '#ffffcc',
                      border: '1px solid #ccc',
                      cursor: 'pointer',
                      '&:hover': { borderColor: '#007bff' },
                    }}
                    title={`Slot: ${slot.ID}\nStatus: ${slot.Status}\nMaterial: ${slot.MaterialID || 'None'}`}
                  >
                    <Typography variant="caption">{slot.Column}</Typography>
                  </Paper>
                </Grid>
              ))}
            </Grid>
          ))}
        </Grid>
      </Box>
    </Box>
  );
}

export default ShelfGrid;
