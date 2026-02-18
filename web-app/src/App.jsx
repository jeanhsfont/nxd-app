import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import MainLayout from './layouts/MainLayout';
import Dashboard from './pages/Dashboard';
import AssetManagement from './pages/AssetManagement';
import Intelligence from './pages/Intelligence';
import Login from './pages/Login';
import Register from './pages/Register';
import Onboarding from './pages/Onboarding';
import Settings from './pages/Settings';
import Sectors from './pages/Sectors';
import IAReports from './pages/IAReports';
import ProtectedRoute from './components/ProtectedRoute';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/onboarding" element={<Onboarding />} />
        <Route path="/dashboard" element={<Dashboard />} />
        
        {/* Rotas Protegidas (Main Layout) */}
        <Route 
          path="/" 
          element={
            <ProtectedRoute>
              <MainLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Dashboard />} />
          <Route path="assets" element={<AssetManagement />} />
          <Route path="ia" element={<Intelligence />} />
          <Route path="ia-reports" element={<IAReports />} />
          <Route path="settings" element={<Settings />} />
          <Route path="sectors" element={<Sectors />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
