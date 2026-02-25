import { HelpCircle, Mail, MessageSquare, FileText } from 'lucide-react';

export default function Support() {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto p-6">
        <div className="page-header">
          <div className="page-header-icon">
            <HelpCircle className="w-6 h-6" />
          </div>
          <div>
            <h1 className="page-title">Central de Suporte</h1>
            <p className="page-subtitle">Estamos aqui para ajudar você</p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <div className="nxd-card hover:border-navy cursor-pointer transition-all">
            <Mail className="w-8 h-8 text-navy mb-4" />
            <h3 className="text-lg font-bold text-gray-900 mb-2">Email</h3>
            <p className="text-gray-600 text-sm mb-3">Entre em contato com nossa equipe de suporte</p>
            <a href="mailto:suporte@nxd.com" className="text-navy hover:underline font-medium text-sm">
              suporte@nxd.com
            </a>
          </div>

          <div className="nxd-card hover:border-navy cursor-pointer transition-all">
            <MessageSquare className="w-8 h-8 text-navy mb-4" />
            <h3 className="text-lg font-bold text-gray-900 mb-2">Chat Online</h3>
            <p className="text-gray-600 text-sm mb-3">Converse com um de nossos especialistas</p>
            <button className="text-navy hover:underline font-medium text-sm">
              Iniciar Chat →
            </button>
          </div>
        </div>

        <div className="nxd-card">
          <div className="flex items-center gap-3 mb-6">
            <FileText className="w-6 h-6 text-navy" />
            <h2 className="text-xl font-bold text-gray-900">Perguntas Frequentes</h2>
          </div>

          <div className="space-y-4">
            <details className="group">
              <summary className="flex items-center justify-between cursor-pointer p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
                <span className="font-medium text-gray-900">Como configurar minha API Key?</span>
                <span className="text-gray-400 group-open:rotate-180 transition-transform">▼</span>
              </summary>
              <div className="mt-3 p-4 text-gray-600 text-sm">
                Acesse Configurações → API Keys e copie sua chave. Use-a no DX Simulator para enviar telemetria.
              </div>
            </details>

            <details className="group">
              <summary className="flex items-center justify-between cursor-pointer p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
                <span className="font-medium text-gray-900">Como adicionar novos ativos?</span>
                <span className="text-gray-400 group-open:rotate-180 transition-transform">▼</span>
              </summary>
              <div className="mt-3 p-4 text-gray-600 text-sm">
                Vá em Gestão de Ativos e clique em "Adicionar Ativo". Preencha as informações e associe ao setor desejado.
              </div>
            </details>

            <details className="group">
              <summary className="flex items-center justify-between cursor-pointer p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
                <span className="font-medium text-gray-900">Como funciona a IA do NXD?</span>
                <span className="text-gray-400 group-open:rotate-180 transition-transform">▼</span>
              </summary>
              <div className="mt-3 p-4 text-gray-600 text-sm">
                Nossa IA analisa dados de telemetria em tempo real e gera insights sobre eficiência, manutenção preventiva e otimizações.
              </div>
            </details>
          </div>
        </div>
      </div>
    </div>
  );
}
