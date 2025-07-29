import React, { useState, useEffect } from 'react';
import { AppBar, Toolbar, Typography, Button, Box, Container, Snackbar, Alert } from '@mui/material';
import { Link } from 'react-router-dom';
import { useSelector, useDispatch } from 'react-redux';
import { addNotification } from '../../store/slices/notificationSlice';

function Layout({ children }) {
  const dispatch = useDispatch();
  const notifications = useSelector((state) => state.notification.messages);
  const [openSnackbar, setOpenSnackbar] = useState(false);
  const [currentNotification, setCurrentNotification] = useState(null);

  useEffect(() => {
    if (notifications.length > 0) {
      setCurrentNotification(notifications[0]);
      setOpenSnackbar(true);
    }
  }, [notifications]);

  const handleCloseSnackbar = (event, reason) => {
    if (reason === 'clickaway') {
      return;
    }
    setOpenSnackbar(false);
    // Optionally remove the displayed notification from the Redux store after it closes
    // dispatch(removeNotification(currentNotification.id));
  };

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            WMS Admin Dashboard
          </Typography>
          <Button color="inherit" component={Link} to="/">
            Dashboard
          </Button>
          <Button color="inherit" component={Link} to="/shelf-management">
            Shelf Management
          </Button>
          <Button color="inherit" component={Link} to="/reports">
            Reports
          </Button>
          <Button color="inherit" component={Link} to="/system-health">
            System Health
          </Button>
        </Toolbar>
      </AppBar>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        {children}
      </Container>
      <Box component="footer" sx={{ bgcolor: 'primary.main', color: 'white', p: 2, textAlign: 'center', width: '100%', position: 'fixed', bottom: 0, left: 0, zIndex: 1000 }}>
        <Typography variant="body2">
          &copy; 2025 WMS. All rights reserved.
        </Typography>
      </Box>
      {currentNotification && (
        <Snackbar open={openSnackbar} autoHideDuration={6000} onClose={handleCloseSnackbar}>
          <Alert onClose={handleCloseSnackbar} severity="info" sx={{ width: '100%' }}>
            {currentNotification.text}
          </Alert>
        </Snackbar>
      )}
    </Box>
  );
}

export default Layout;
