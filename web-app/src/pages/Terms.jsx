import { FileText, ArrowLeft } from 'lucide-react';
import { Link } from 'react-router-dom';

export default function Terms() {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto p-6">
        <Link to="/login" className="inline-flex items-center gap-2 text-gray-600 hover:text-navy mb-6 transition-colors">
          <ArrowLeft className="w-4 h-4" />
          Voltar ao Login
        </Link>

        <div className="nxd-card">
          <div className="flex items-center gap-3 mb-6">
            <FileText className="w-8 h-8 text-navy" />
            <h1 className="text-3xl font-bold text-gray-900">Termos de Uso</h1>
          </div>

          <div className="prose max-w-none">
            <p className="text-gray-600 mb-6">
              Última atualização: {new Date().toLocaleDateString('pt-BR')}
            </p>

            <section className="mb-8">
              <h2 className="text-xl font-bold text-gray-900 mb-3">1. Aceitação dos Termos</h2>
              <p className="text-gray-600">
                Ao acessar e usar a plataforma NXD (Nexus Data Exchange), você concorda com estes Termos de Uso.
                Se você não concorda com qualquer parte destes termos, não deve usar nossa plataforma.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-bold text-gray-900 mb-3">2. Uso da Plataforma</h2>
              <p className="text-gray-600 mb-3">
                A plataforma NXD é destinada ao monitoramento e análise de dados industriais em tempo real.
                Você concorda em:
              </p>
              <ul className="list-disc pl-6 text-gray-600 space-y-2">
                <li>Fornecer informações precisas e atualizadas durante o cadastro</li>
                <li>Manter a confidencialidade de suas credenciais de acesso</li>
                <li>Não compartilhar sua API Key com terceiros não autorizados</li>
                <li>Usar a plataforma apenas para fins legítimos e industriais</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-bold text-gray-900 mb-3">3. Privacidade e Dados</h2>
              <p className="text-gray-600">
                Respeitamos sua privacidade. Todos os dados de telemetria são armazenados com segurança e
                utilizados apenas para fornecer os serviços da plataforma. Não compartilhamos seus dados
                com terceiros sem seu consentimento explícito.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-bold text-gray-900 mb-3">4. Propriedade Intelectual</h2>
              <p className="text-gray-600">
                Todo o conteúdo, design e funcionalidades da plataforma NXD são de propriedade exclusiva
                da NXD Technologies e estão protegidos por leis de direitos autorais.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-bold text-gray-900 mb-3">5. Limitação de Responsabilidade</h2>
              <p className="text-gray-600">
                A plataforma NXD é fornecida "como está". Não garantimos que o serviço será ininterrupto
                ou livre de erros. Não nos responsabilizamos por danos diretos ou indiretos resultantes
                do uso da plataforma.
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-3">6. Contato</h2>
              <p className="text-gray-600">
                Para questões sobre estes Termos de Uso, entre em contato conosco através do email:
                <a href="mailto:legal@nxd.com" className="text-navy hover:underline ml-1">legal@nxd.com</a>
              </p>
            </section>
          </div>
        </div>
      </div>
    </div>
  );
}
