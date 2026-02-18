import React, { useState } from 'react';
import ApiKeyModal from '../components/ApiKeyModal';

export default function Onboarding() {
  const [step, setStep] = useState(1);
  const [personalData, setPersonalData] = useState({ fullName: '', cpf: '' });
  const [factoryData, setFactoryData] = useState({ name: '', cnpj: '', address: '' });
  const [isTwoFactorEnabled, setIsTwoFactorEnabled] = useState(false);
  const [showKeyModal, setShowKeyModal] = useState(false);
  const [generatedApiKey, setGeneratedApiKey] = useState('');

  const handlePersonalChange = (e) => setPersonalData({ ...personalData, [e.target.name]: e.target.value });
  const handleFactoryChange = (e) => setFactoryData({ ...factoryData, [e.target.name]: e.target.value });

  const handleSubmit = async () => {
    try {
      const token = localStorage.getItem('nxd-token');
      const response = await fetch('/api/onboarding', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ personalData, factoryData, twoFactorEnabled: isTwoFactorEnabled }),
      });

      if (!response.ok) {
        throw new Error('Falha ao finalizar o cadastro.');
      }

      const data = await response.json();
      setGeneratedApiKey(data.apiKey);
      setShowKeyModal(true);
    } catch (error) {
      console.error(error);
    }
  };

  const handleCloseModal = () => {
    // Após fechar o modal, redireciona para o sistema principal
    window.location.href = '/'; 
  };

  const handleNextStep = () => setStep(step + 1);
  const handlePrevStep = () => setStep(step - 1);

  return (
    <>
      {showKeyModal && <ApiKeyModal apiKey={generatedApiKey} onClose={handleCloseModal} />}
      <div className={`flex items-center justify-center min-h-screen bg-gray-100 ${showKeyModal ? 'filter blur-sm' : ''}`}>
        <div className="p-8 bg-white rounded-lg shadow-md w-full max-w-lg">
          <h1 className="text-2xl font-bold mb-4 text-center">Finalize seu Cadastro</h1>
          {
            {
              1: (
                <div>
                  <h2 className="text-xl font-semibold mb-4">Passo 1: Seus Dados Pessoais</h2>
                  <div className="mb-4">
                    <label className="block text-gray-700 font-bold mb-2">Nome Completo</label>
                    <input type="text" name="fullName" value={personalData.fullName} onChange={handlePersonalChange} className="w-full px-3 py-2 border rounded-lg" />
                  </div>
                  <div className="mb-4">
                    <label className="block text-gray-700 font-bold mb-2">CPF</label>
                    <input type="text" name="cpf" value={personalData.cpf} onChange={handlePersonalChange} className="w-full px-3 py-2 border rounded-lg" />
                  </div>
                  <button onClick={handleNextStep} className="w-full bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg">
                    Próximo
                  </button>
                </div>
              ),
              2: (
                <div>
                  <h2 className="text-xl font-semibold mb-4">Passo 2: Dados da sua Fábrica</h2>
                  <div className="mb-4">
                    <label className="block text-gray-700 font-bold mb-2">Nome da Fábrica</label>
                    <input type="text" name="name" value={factoryData.name} onChange={handleFactoryChange} className="w-full px-3 py-2 border rounded-lg" />
                  </div>
                  <div className="mb-4">
                    <label className="block text-gray-700 font-bold mb-2">CNPJ</label>
                    <input type="text" name="cnpj" value={factoryData.cnpj} onChange={handleFactoryChange} className="w-full px-3 py-2 border rounded-lg" />
                  </div>
                  <div className="mb-4">
                    <label className="block text-gray-700 font-bold mb-2">Endereço</label>
                    <input type="text" name="address" value={factoryData.address} onChange={handleFactoryChange} className="w-full px-3 py-2 border rounded-lg" />
                  </div>
                  <div className="flex justify-between">
                    <button onClick={handlePrevStep} className="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded-lg">
                      Anterior
                    </button>
                    <button onClick={handleNextStep} className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg">
                      Próximo
                    </button>
                  </div>
                </div>
              ),
              3: (
                <div>
                  <h2 className="text-xl font-semibold mb-4">Passo 3: Segurança da Conta</h2>
                  <p className="text-gray-600 mb-4">
                    Para garantir a integridade dos seus dados e liberar a geração da chave de API, é obrigatório ativar a autenticação de dois fatores (2FA).
                  </p>
                  <div className="p-4 bg-gray-100 rounded-lg mb-4">
                    <p className="text-center font-semibold">[Simulação de Ativação de 2FA]</p>
                  </div>
                  <div className="flex items-center mb-6">
                    <input type="checkbox" id="2fa-check" className="mr-2" checked={isTwoFactorEnabled} onChange={(e) => setIsTwoFactorEnabled(e.target.checked)} />
                    <label htmlFor="2fa-check">Li e ativei a autenticação de dois fatores.</label>
                  </div>
                  <div className="flex justify-between">
                    <button onClick={handlePrevStep} className="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded-lg">
                      Anterior
                    </button>
                    <button onClick={handleSubmit} disabled={!isTwoFactorEnabled} className="w-1/2 bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded-lg disabled:bg-gray-400">
                      Finalizar Cadastro e Gerar Chave
                    </button>
                  </div>
                </div>
              )
            }[step]
          }
        </div>
      </div>
    </>
  );
}
