import React, { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation, Link } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import {
  LayoutDashboard,
  Factory,
  Package,
  DollarSign,
  Brain,
  FileText,
  Settings,
  CreditCard,
  HelpCircle,
  LogOut,
  Menu,
  X,
  Zap
} from 'lucide-react';
import './App.css';

// Pages
import Login from './pages/Login';
import Register from './pages/Register';
import Onboarding from './pages/Onboarding';
import Welcome from './pages/Welcome';
import Dashboard from './pages/Dashboard';
import AssetManagement from './pages/AssetManagement';
import FinancialIndicators from './pages/FinancialIndicators';
import Sectors from './pages/Sectors';
import Intelligence from './pages/Intelligence';
import IAReports from './pages/IAReports';
import SettingsPage from './pages/Settings';
import Billing from './pages/Billing';
import Support from './pages/Support';
import Terms from './pages/Terms';
import NotFound from './pages/NotFound';

function ProtectedRoute({ children }) {
  const token = localStorage.getItem('nxd-token');
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  return children;
}

const navItems = [
  { path: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { path: '/assets', icon: Factory, label: 'Gestão de Ativos' },
  { path: '/sectors', icon: Package, label: 'Setores' },
  { path: '/financial', icon: DollarSign, label: 'Financeiro' },
  { path: '/intelligence', icon: Brain, label: 'Intelligence' },
  { path: '/reports', icon: FileText, label: 'Relatórios IA' },
  { path: '/settings', icon: Settings, label: 'Configurações' },
  { path: '/billing', icon: CreditCard, label: 'Cobrança' },
  { path: '/support', icon: HelpCircle, label: 'Suporte' },
];

function NavItem({ item, isActive }) {
  const Icon = item.icon;
  
  return (
    <Link to={item.path} className={`nav-item ${isActive ? 'active' : ''}`}>
      <Icon className="w-5 h-5" />
      <span>{item.label}</span>
    </Link>
  );
}

function Sidebar({ isOpen, onClose }) {
  const location = useLocation();
  
  const handleLogout = () => {
    localStorage.removeItem('nxd-token');
    window.location.href = '/login';
  };
  
  return (
    <>
      {/* Mobile overlay */}
      {isOpen && (
        <div
          onClick={onClose}
          className="fixed inset-0 bg-black/20 z-40 lg:hidden"
        />
      )}
      
      {/* Sidebar */}
      <aside className={`fixed lg:static top-0 left-0 z-50 sidebar ${
        isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
      } transition-transform duration-300 ease-in-out`}>
        
        {/* Header */}
        <div className="sidebar-header">
          <div className="flex items-center justify-between">
            <div className="sidebar-logo">
              <div className="sidebar-logo-icon">
                <Zap className="w-5 h-5" />
              </div>
              <div>
                <h1 className="text-lg font-bold text-gray-900">NXD</h1>
                <p className="text-xs text-gray-500">Nexus Data Exchange</p>
              </div>
            </div>
            
            <button
              onClick={onClose}
              className="lg:hidden text-gray-400 hover:text-gray-600"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>
        
        {/* Navigation */}
        <nav className="sidebar-nav">
          {navItems.map((item) => (
            <NavItem
              key={item.path}
              item={item}
              isActive={location.pathname === item.path}
            />
          ))}
        </nav>
        
        {/* Footer */}
        <div className="p-4 border-t border-gray-200">
          <button
            onClick={handleLogout}
            className="w-full flex items-center gap-3 px-4 py-3 text-red-600 hover:bg-red-50 rounded-lg transition-colors text-sm font-medium"
          >
            <LogOut className="w-5 h-5" />
            <span>Sair</span>
          </button>
        </div>
      </aside>
    </>
  );
}

function AppLayout({ children }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const location = useLocation();
  
  useEffect(() => {
    setSidebarOpen(false);
  }, [location.pathname]);
  
  return (
    <div className="flex h-screen bg-gray-50">
      <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
      
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Mobile header */}
        <header className="lg:hidden bg-white border-b border-gray-200 px-6 py-4">
          <div className="flex items-center justify-between">
            <button
              onClick={() => setSidebarOpen(true)}
              className="text-gray-600"
            >
              <Menu className="w-6 h-6" />
            </button>
            
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-navy rounded-lg flex items-center justify-center">
                <Zap className="w-5 h-5 text-white" />
              </div>
              <span className="font-bold text-gray-900">NXD</span>
            </div>
            
            <div className="w-6"></div>
          </div>
        </header>
        
        {/* Main content */}
        <main className="flex-1 overflow-auto">
          {children}
        </main>
      </div>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <Toaster
        position="top-right"
        toastOptions={{
          duration: 4000,
          style: {
            background: '#ffffff',
            color: '#111827',
            border: '1px solid #e5e7eb',
            borderRadius: '8px',
            padding: '16px',
            boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
          },
          success: {
            iconTheme: {
              primary: '#16a34a',
              secondary: '#ffffff',
            },
          },
          error: {
            iconTheme: {
              primary: '#dc2626',
              secondary: '#ffffff',
            },
          },
        }}
      />
      
      <Routes>
        {/* Public routes */}
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/terms" element={<Terms />} />
        <Route path="/support" element={<Support />} />
        
        {/* Protected routes */}
        <Route path="/onboarding" element={<ProtectedRoute><Onboarding /></ProtectedRoute>} />
        <Route path="/welcome" element={<ProtectedRoute><Welcome /></ProtectedRoute>} />
        
        {/* App routes with layout */}
        <Route path="/" element={<ProtectedRoute><AppLayout><Dashboard /></AppLayout></ProtectedRoute>} />
        <Route path="/assets" element={<ProtectedRoute><AppLayout><AssetManagement /></AppLayout></ProtectedRoute>} />
        <Route path="/sectors" element={<ProtectedRoute><AppLayout><Sectors /></AppLayout></ProtectedRoute>} />
        <Route path="/financial" element={<ProtectedRoute><AppLayout><FinancialIndicators /></AppLayout></ProtectedRoute>} />
        <Route path="/intelligence" element={<ProtectedRoute><AppLayout><Intelligence /></AppLayout></ProtectedRoute>} />
        <Route path="/reports" element={<ProtectedRoute><AppLayout><IAReports /></AppLayout></ProtectedRoute>} />
        <Route path="/settings" element={<ProtectedRoute><AppLayout><SettingsPage /></AppLayout></ProtectedRoute>} />
        <Route path="/billing" element={<ProtectedRoute><AppLayout><Billing /></AppLayout></ProtectedRoute>} />
        
        {/* 404 */}
        <Route path="*" element={<NotFound />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
