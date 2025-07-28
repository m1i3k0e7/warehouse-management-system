import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Layout from './components/common/Layout';
import Dashboard from './pages/Dashboard';
import ShelfView from './pages/ShelfView';
import Profile from './pages/Profile';
import MaterialOperations from './pages/MaterialOperations';

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/shelf/:shelfId" element={<ShelfView />} />
          <Route path="/operations" element={<MaterialOperations />} />
          <Route path="/profile" element={<Profile />} />
        </Routes>
      </Layout>
    </Router>
  );
}

export default App;
