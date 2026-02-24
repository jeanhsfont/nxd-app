import React from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Cuboid, Bot, Settings, LogOut, Download, DollarSign, CreditCard, FileText, HelpCircle } from 'lucide-react';
import { clsx } from 'clsx';

function NavItem({ to, icon: Icon, children }) {
  const location = useLocation();
  const isActive = location.pathname === to;
  
  return (
    <Link 
      to={to} 
      className={clsx(
        "flex items-center gap-3 px-4 py-3 rounded-lg transition-colors font-medium",
        isActive 
          ? "bg-indigo-50 text-indigo-700" 
          : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
      )}
    >
      <Icon className="w-5 h-5" />
      {children}
    </Link>
  );
}

export default function MainLayout({ contentOverride }) {
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    localStorage.removeItem('nxd-token');
    navigate('/login');
  };

  return (
    <div className="flex h-screen bg-gray-50 font-sans">
      {/* Global Sidebar */}
      <aside className="w-64 bg-white border-r border-gray-200 flex flex-col fixed h-full z-20">
        <div className="p-6 border-b border-gray-100">
          <div className="flex items-center gap-2 text-indigo-600 font-bold text-xl">
            <Cuboid className="w-8 h-8" />
            <span>NXD v2.0</span>
          </div>
        </div>

        <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
          <div className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2 mt-4 px-4">
            Operação
          </div>
          <NavItem to="/" icon={LayoutDashboard}>Dashboard</NavItem>
          <NavItem to="/assets" icon={Cuboid}>Gestão de Ativos</NavItem>
          
          <div className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2 mt-8 px-4">
            Inteligência
          </div>
          <NavItem to="/ia" icon={Bot}>NXD Intelligence</NavItem>
          <NavItem to="/import" icon={Download}>Importar histórico</NavItem>
          <NavItem to="/financial" icon={DollarSign}>Indicadores financeiros</NavItem>

          <div className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2 mt-8 px-4">
            Configuração
          </div>
          <NavItem to="/settings" icon={Settings}>Ajustes</NavItem>
          <NavItem to="/billing" icon={CreditCard}>Cobrança</NavItem>
          <NavItem to="/terms" icon={FileText}>Termos</NavItem>
          <NavItem to="/support" icon={HelpCircle}>Suporte</NavItem>
        </nav>

        <div className="p-4 border-t border-gray-100">
          <button onClick={handleLogout} className="flex items-center gap-3 px-4 py-3 w-full text-left text-red-600 hover:bg-red-50 rounded-lg transition-colors font-medium">
            <LogOut className="w-5 h-5" />
            Sair
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 ml-64 overflow-auto">
        {contentOverride ?? <Outlet />}
      </div>
    </div>
  );
}
