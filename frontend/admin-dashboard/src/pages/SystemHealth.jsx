import React, { useState, useEffect } from 'react';
import { Container, Typography, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip } from '@mui/material';
import { inventoryApi } from '../services/api';

function createData(service, status) {
  return { service, status };
}

function SystemHealth() {
  const [healthData, setHealthData] = useState([]);

  useEffect(() => {
    const fetchHealthData = async () => {
      try {
        // Replace with your actual API call to get system health data
        // const response = await inventoryApi.getSystemHealth();
        const mockHealthData = [
          createData('Inventory Service', 'Healthy'),
          createData('Realtime Service', 'Healthy'),
          createData('Location Service', 'Unhealthy'),
          createData('Analytics Service', 'Healthy'),
          createData('Notification Service', 'Healthy'),
          createData('PostgreSQL Database', 'Healthy'),
          createData('Redis Cache', 'Healthy'),
          createData('Kafka Broker', 'Healthy'),
        ];
        setHealthData(mockHealthData);
      } catch (error) {
        console.error('Failed to fetch system health data:', error);
      }
    };

    fetchHealthData();
  }, []);

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        System Health
      </Typography>
      <TableContainer component={Paper}>
        <Table sx={{ minWidth: 650 }} aria-label="simple table">
          <TableHead>
            <TableRow>
              <TableCell>Service</TableCell>
              <TableCell align="right">Status</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {healthData.map((row) => (
              <TableRow
                key={row.service}
                sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
              >
                <TableCell component="th" scope="row">
                  {row.service}
                </TableCell>
                <TableCell align="right">
                  <Chip
                    label={row.status}
                    color={row.status === 'Healthy' ? 'success' : 'error'}
                  />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Container>
  );
}

export default SystemHealth;
