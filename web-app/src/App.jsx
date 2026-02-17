import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import MainLayout from './layouts/MainLayout';
import Dashboard from './pages/Dashboard';
import AssetManagement from './pages/AssetManagement';
import Intelligence from './pages/Intelligence';
import Login from './pages/Login';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        
        {/* Rotas Protegidas (Main Layout) */}
        <Route path="/" element={<MainLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="assets" element={<AssetManagement />} />
          <Route path="ia" element={<Intelligence />} />
          <Route path="settings" element={<div className="p-8">Configurações (Em breve)</div>} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
