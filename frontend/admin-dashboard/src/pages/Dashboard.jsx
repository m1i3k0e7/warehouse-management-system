import React, { useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { Container, Grid, Typography } from '@mui/material';
import StatsCard from '../components/dashboard/StatsCard';
import RecentOperations from '../components/dashboard/RecentOperations';
import { fetchStatsStart, fetchStatsSuccess, fetchStatsFailure } from '../store/slices/dashboardSlice';
import { inventoryApi } from '../services/api'; // Assuming you have an API call for stats
import StorageIcon from '@mui/icons-material/Storage';
import CheckBoxOutlineBlankIcon from '@mui/icons-material/CheckBoxOutlineBlank';
import CheckBoxIcon from '@mui/icons-material/CheckBox';
import DnsIcon from '@mui/icons-material/Dns';

function Dashboard() {
  const dispatch = useDispatch();
  const { stats, loading, error } = useSelector((state) => state.dashboard);

  useEffect(() => {
    const fetchStats = async () => {
      dispatch(fetchStatsStart());
      try {
        const response = await inventoryApi.getDashboardStats();
        dispatch(fetchStatsSuccess(response.data));
      } catch (err) {
        dispatch(fetchStatsFailure(err.message));
      }
    };

    fetchStats();
  }, [dispatch]);

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Admin Dashboard
      </Typography>
      <Grid container spacing={3}>
        <Grid item xs={12} sm={6} md={3}>
          <StatsCard title="Total Shelves" value={stats.totalShelves} icon={<DnsIcon sx={{ fontSize: 40 }} />} />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatsCard title="Total Slots" value={stats.totalSlots} icon={<StorageIcon sx={{ fontSize: 40 }} />} />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatsCard title="Occupied Slots" value={stats.occupiedSlots} icon={<CheckBoxIcon sx={{ fontSize: 40 }} />} />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatsCard title="Empty Slots" value={stats.emptySlots} icon={<CheckBoxOutlineBlankIcon sx={{ fontSize: 40 }} />} />
        </Grid>
        <Grid item xs={12} md={8}>
          {/* Placeholder for a chart */}
          <Typography>Charts and graphs will be here.</Typography>
        </Grid>
        <Grid item xs={12} md={4}>
          <RecentOperations />
        </Grid>
      </Grid>
    </Container>
  );
}

export default Dashboard;
