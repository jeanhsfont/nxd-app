import React from 'react';
import { FileText } from 'lucide-react';

/**
 * Termos de Uso do NXD (Nexus Data Exchange).
 * Documento vinculante; em produção manter atualizado e revisado por jurídico.
 */
export default function Terms() {
  return (
    <div className="p-8 max-w-3xl mx-auto">
      <div className="flex items-center gap-2 mb-6">
        <FileText className="w-8 h-8 text-indigo-600" />
        <h1 className="text-2xl font-bold text-gray-900">Termos de Uso</h1>
      </div>

      <div className="bg-white border border-gray-200 rounded-xl shadow-sm p-6 prose prose-sm max-w-none text-gray-700">
        <p className="text-gray-500 text-xs uppercase tracking-wide mb-4">
          Última atualização: Fevereiro de 2026
        </p>

        <h2 className="text-lg font-semibold text-gray-900 mt-6">1. Aceitação</h2>
        <p>
          Ao acessar, registrar-se ou utilizar o serviço NXD (Nexus Data Exchange) e quaisquer
          funcionalidades associadas, você concorda integralmente com estes Termos de Uso. Se você
          não concordar com qualquer parte destes termos, não utilize o serviço.
        </p>

        <h2 className="text-lg font-semibold text-gray-900 mt-6">2. Descrição do Serviço</h2>
        <p>
          O NXD é uma plataforma de monitoramento e troca de dados industriais que permite conectar
          máquinas, CLPs e sistemas de fábrica a um painel centralizado. O serviço inclui ingestão
          de dados em tempo real, armazenamento de histórico, relatórios, indicadores financeiros e
          análise assistida por IA, conforme os planos e recursos disponibilizados na sua conta.
        </p>

        <h2 className="text-lg font-semibold text-gray-900 mt-6">3. Cadastro e Conta</h2>
        <p>
          É necessário cadastro com e-mail e senha válidos. Você é responsável por manter a
          confidencialidade das credenciais e pela atividade realizada em sua conta. A API Key
          fornecida pelo NXD é de uso exclusivo do titular da conta e não deve ser compartilhada
          com terceiros. O NXD reserva-se o direito de suspender ou encerrar contas que violem
          estes termos ou que sejam utilizadas de forma fraudulenta ou abusiva.
        </p>

        <h2 className="text-lg font-semibold text-gray-900 mt-6">4. Uso Aceitável</h2>
        <p>
          O uso do NXD deve ser lícito e em conformidade com a legislação aplicável. É proibido:
          utilizar o serviço para fins ilícitos; sobrecarregar a infraestrutura de forma
          intencional; tentar acessar dados de outros usuários; descompilar, fazer engenharia
          reversa ou tentar extrair o código-fonte do serviço; ou utilizar bots ou automação não
          autorizada além das integrações documentadas (por exemplo, envio de dados via API).
        </p>

        <h2 className="text-lg font-semibold text-gray-900 mt-6">5. Dados e Privacidade</h2>
        <p>
          Os dados de telemetria e operacionais que você envia ao NXD são processados e armazenados
          para fins de operação do serviço, geração de relatórios e melhoria do produto. O NXD
          atua como processador dos seus dados conforme a Política de Privacidade em vigor. Você
          garante que possui direito legal para enviar e processar esses dados no NXD e que o
          tratamento está em conformidade com a LGPD e demais normas aplicáveis ao seu negócio.
        </p>

        <h2 className="text-lg font-semibold text-gray-900 mt-6">6. Cobrança e Planos</h2>
        <p>
          Planos pagos estão sujeitos aos preços e condições vigentes no momento da assinatura.
          Renovações e cobranças serão realizadas conforme o método de pagamento cadastrado. O
          cancelamento pode ser feito conforme as instruções na área de Cobrança; o acesso
          permanece até o fim do período já pago. Reembolsos estão sujeitos à política de
          reembolso publicada no site ou comunicada ao usuário.
        </p>

        <h2 className="text-lg font-semibold text-gray-900 mt-6">7. Disponibilidade e Suporte</h2>
        <p>
          O NXD busca manter alta disponibilidade; eventuais janelas de manutenção serão
          comunicadas quando possível. O suporte é oferecido conforme o plano contratado (e-mail,
          documentação e, em planos superiores, canais adicionais). Não há garantia de resultado
          específico ou de que o serviço estará livre de interrupções.
        </p>

        <h2 className="text-lg font-semibold text-gray-900 mt-6">8. Limitação de Responsabilidade</h2>
        <p>
          Na máxima extensão permitida pela lei, o NXD e seus fornecedores não se responsabilizam
          por danos indiretos, incidentais, especiais ou consequenciais decorrentes do uso ou da
          impossibilidade de uso do serviço. A responsabilidade total não excederá o valor pago
          pelo usuário nos doze meses anteriores ao fato que deu causa ao dano.
        </p>

        <h2 className="text-lg font-semibold text-gray-900 mt-6">9. Alterações</h2>
        <p>
          O NXD pode alterar estes Termos de Uso a qualquer momento. Alterações relevantes serão
          comunicadas por e-mail ou aviso no produto. O uso continuado do serviço após a vigência
          das alterações constitui aceitação dos novos termos.
        </p>

        <h2 className="text-lg font-semibold text-gray-900 mt-6">10. Contato</h2>
        <p>
          Dúvidas sobre estes termos ou sobre o serviço: utilize a área de Suporte no sistema
          ou entre em contato pelo e-mail de suporte indicado no produto (ex.: suporte@nxd.io).
        </p>
      </div>
    </div>
  );
}
