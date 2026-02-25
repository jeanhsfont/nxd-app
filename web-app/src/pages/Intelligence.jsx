import { Brain } from 'lucide-react';
import Chat from '../components/Chat';

export default function Intelligence() {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-5xl mx-auto p-6">
        <div className="page-header">
          <div className="page-header-icon">
            <Brain className="w-6 h-6" />
          </div>
          <div>
            <h1 className="page-title">NXD Intelligence</h1>
            <p className="page-subtitle">Converse com IA sobre seus dados industriais</p>
          </div>
        </div>

        <div className="nxd-card">
          <Chat />
        </div>
      </div>
    </div>
  );
}
