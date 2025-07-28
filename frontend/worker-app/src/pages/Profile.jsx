import React from 'react';
import { useSelector } from 'react-redux';
import { Container, Typography, Paper, Box, Button } from '@mui/material';
import { useDispatch } from 'react-redux';
import { logout } from '../store/slices/userSlice';

function Profile() {
  const dispatch = useDispatch();
  const { userId, role, isAuthenticated } = useSelector((state) => state.user);

  const handleLogout = () => {
    dispatch(logout());
    // Redirect to login page or home page after logout
    // navigate('/login');
  };

  return (
    <Container maxWidth="sm" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        User Profile
      </Typography>

      <Paper elevation={3} sx={{ p: 3 }}>
        {!isAuthenticated ? (
          <Typography variant="h6" color="error">Not Authenticated</Typography>
        ) : (
          <Box>
            <Typography variant="body1"><strong>User ID:</strong> {userId}</Typography>
            <Typography variant="body1"><strong>Role:</strong> {role}</Typography>
            <Button variant="contained" color="secondary" onClick={handleLogout} sx={{ mt: 3 }}>
              Logout
            </Button>
          </Box>
        )}
      </Paper>
    </Container>
  );
}

export default Profile;
