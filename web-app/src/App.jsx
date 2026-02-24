import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { getPageTitle } from './config/titles';
import MainLayout from './layouts/MainLayout';
import Dashboard from './pages/Dashboard';
import AssetManagement from './pages/AssetManagement';
import Intelligence from './pages/Intelligence';
import Login from './pages/Login';
import Register from './pages/Register';
import Onboarding from './pages/Onboarding';
import Settings from './pages/Settings';
import ImportHistoric from './pages/ImportHistoric';
import FinancialIndicators from './pages/FinancialIndicators';
import Billing from './pages/Billing';
import Terms from './pages/Terms';
import Support from './pages/Support';
import NotFound from './pages/NotFound';
import Welcome from './pages/Welcome';
import ProtectedRoute from './components/ProtectedRoute';

// Autenticado: Terms/Support dentro do MainLayout. Não autenticado: layout simples.
function TermsRoute() {
  const hasToken = typeof window !== 'undefined' && !!localStorage.getItem('nxd-token');
  if (hasToken) {
    return (
      <ProtectedRoute>
        <MainLayout contentOverride={<Terms />} />
      </ProtectedRoute>
    );
  }
  return <Terms />;
}

function SupportRoute() {
  const hasToken = typeof window !== 'undefined' && !!localStorage.getItem('nxd-token');
  if (hasToken) {
    return (
      <ProtectedRoute>
        <MainLayout contentOverride={<Support />} />
      </ProtectedRoute>
    );
  }
  return <Support />;
}

function DocumentTitle() {
  const { pathname } = useLocation();
  useEffect(() => {
    document.title = getPageTitle(pathname);
    return () => { document.title = 'NXD'; };
  }, [pathname]);
  return null;
}

function App() {
  return (
    <BrowserRouter>
      <DocumentTitle />
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/terms" element={<TermsRoute />} />
        <Route path="/support" element={<SupportRoute />} />
        <Route path="/onboarding" element={<Onboarding />} />

        {/* Tela de conexão (protegida, sem sidebar) */}
        <Route path="/welcome" element={<ProtectedRoute><Welcome /></ProtectedRoute>} />

        {/* Rotas Protegidas (Main Layout). /dashboard redireciona para / */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <MainLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Dashboard />} />
          <Route path="dashboard" element={<Navigate to="/" replace />} />
          <Route path="assets" element={<AssetManagement />} />
          <Route path="ia" element={<Intelligence />} />
          <Route path="import" element={<ImportHistoric />} />
          <Route path="financial" element={<FinancialIndicators />} />
          <Route path="settings" element={<Settings />} />
          <Route path="billing" element={<Billing />} />
        </Route>

        {/* 404 — qualquer rota não declarada */}
        <Route path="*" element={<NotFound />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
