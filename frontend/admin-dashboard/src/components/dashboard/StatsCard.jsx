import React from 'react';
import { Card, CardContent, Typography, Box } from '@mui/material';

function StatsCard({ title, value, icon }) {
  return (
    <Card sx={{ display: 'flex', alignItems: 'center', p: 2 }}>
      <Box sx={{ flexGrow: 1 }}>
        <Typography color="text.secondary" gutterBottom>
          {title}
        </Typography>
        <Typography variant="h4" component="div">
          {value}
        </Typography>
      </Box>
      <Box>
        {icon}
      </Box>
    </Card>
  );
}

export default StatsCard;
