import React from 'react';
import { useSelector } from 'react-redux';
import { List, ListItem, ListItemText, Typography, Paper } from '@mui/material';

function RecentOperations() {
  const operationHistory = useSelector((state) => state.operation.operationHistory);

  return (
    <Paper elevation={3} sx={{ p: 2 }}>
      <Typography variant="h6" gutterBottom>
        Recent Operations
      </Typography>
      <List>
        {operationHistory.length === 0 ? (
          <ListItem>
            <ListItemText primary="No recent operations." />
          </ListItem>
        ) : (
          operationHistory.slice(0, 10).map((op, index) => (
            <ListItem key={index} divider>
              <ListItemText
                primary={`${op.type.replace(/_/g, ' ').toUpperCase()}`}
                secondary={`By ${op.operatorId} at ${new Date(op.timestamp).toLocaleTimeString()}`}
              />
            </ListItem>
          ))
        )}
      </List>
    </Paper>
  );
}

export default RecentOperations;
