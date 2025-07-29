import React, { useState, useEffect } from 'react';
import { Container, Typography, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import { inventoryApi } from '../services/api';

function createData(name, value) {
  return { name, value };
}

function Reports() {
  const [reportData, setReportData] = useState([]);

  useEffect(() => {
    const fetchReportData = async () => {
      try {
        // Replace with your actual API call to get report data
        // const response = await inventoryApi.getReports();
        const mockReportData = [
          createData('Inventory Turnover Rate', '12.5'),
          createData('Average Slot Utilization', '85%'),
          createData('Picking Accuracy', '99.8%'),
          createData('Average Order Fulfillment Time', '2.5 hours'),
        ];
        setReportData(mockReportData);
      } catch (error) {
        console.error('Failed to fetch report data:', error);
      }
    };

    fetchReportData();
  }, []);

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Reports
      </Typography>
      <TableContainer component={Paper}>
        <Table sx={{ minWidth: 650 }} aria-label="simple table">
          <TableHead>
            <TableRow>
              <TableCell>Report Name</TableCell>
              <TableCell align="right">Value</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {reportData.map((row) => (
              <TableRow
                key={row.name}
                sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
              >
                <TableCell component="th" scope="row">
                  {row.name}
                </TableCell>
                <TableCell align="right">{row.value}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Container>
  );
}

export default Reports;
