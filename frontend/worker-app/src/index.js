import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import App from './App';
import store from './store';
import theme from './theme'; // Import your custom theme
import './styles/globals.css';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <Provider store={store}>
      <ThemeProvider theme={theme}>
        <CssBaseline /> {/* CssBaseline to kickstart an elegant, consistent, and simple baseline to build upon. */}
        <App />
      </ThemeProvider>
    </Provider>
  </React.StrictMode>
);
