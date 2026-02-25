import { Link } from 'react-router-dom';
import { Home, AlertCircle } from 'lucide-react';

export default function NotFound() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="text-center">
        <div className="inline-flex items-center justify-center w-20 h-20 bg-red/10 rounded-xl mb-6">
          <AlertCircle className="w-10 h-10 text-red" />
        </div>
        <h1 className="text-6xl font-black text-navy mb-4">404</h1>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">Página não encontrada</h2>
        <p className="text-gray-600 mb-8">A página que você está procurando não existe.</p>
        <Link to="/">
          <button className="nxd-btn nxd-btn-primary">
            <Home className="w-5 h-5" />
            Voltar ao Dashboard
          </button>
        </Link>
      </div>
    </div>
  );
}
