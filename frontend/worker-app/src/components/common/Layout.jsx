import React from 'react';
import { AppBar, Toolbar, Typography, Button, Box, Container } from '@mui/material';
import { Link } from 'react-router-dom';

function Layout({ children }) {
  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            WMS Worker App
          </Typography>
          <Button color="inherit" component={Link} to="/">
            Dashboard
          </Button>
          <Button color="inherit" component={Link} to="/operations">
            Operations
          </Button>
          <Button color="inherit" component={Link} to="/profile">
            Profile
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
    </Box>
  );
}

export default Layout;
