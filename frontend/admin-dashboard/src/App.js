import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Layout from './components/common/Layout';
import Dashboard from './pages/Dashboard';
import ShelfManagement from './pages/ShelfManagement';
import Reports from './pages/Reports';
import SystemHealth from './pages/SystemHealth';

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/shelf-management" element={<ShelfManagement />} />
          <Route path="/reports" element={<Reports />} />
          <Route path="/system-health" element={<SystemHealth />} />
        </Routes>
      </Layout>
    </Router>
  );
}

export default App;
