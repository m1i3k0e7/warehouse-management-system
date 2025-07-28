import { createTheme } from '@mui/material/styles';

const theme = createTheme({
  palette: {
    primary: {
      main: '#212121', // Dark Gray
    },
    secondary: {
      main: '#424242', // Medium Gray
    },
    background: {
      default: '#f5f5f5', // Light Gray
      paper: '#ffffff', // White
    },
    text: {
      primary: '#212121',
      secondary: '#616161', // Gray
    },
    error: {
      main: '#d32f2f', // Red for errors
    },
    warning: {
      main: '#fbc02d', // Amber for warnings
    },
    info: {
      main: '#2196f3', // Blue for info
    },
    success: {
      main: '#4caf50', // Green for success
    },
  },
  typography: {
    fontFamily: 'Roboto, Arial, sans-serif',
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          boxShadow: 'none',
          borderBottom: '1px solid #e0e0e0',
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          boxShadow: '0px 2px 4px rgba(0, 0, 0, 0.05)',
        },
      },
    },
  },
});

export default theme;
