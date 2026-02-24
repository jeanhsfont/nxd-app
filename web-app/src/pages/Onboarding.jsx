import React, { useState, useEffect } from 'react';
import ApiKeyModal from '../components/ApiKeyModal';
import api from '../utils/api';

export default function Onboarding() {
  const [step, setStep] = useState(1);
  const [personalData, setPersonalData] = useState({ fullName: '', cpf: '' });
  const [factoryData, setFactoryData] = useState({ name: '', cnpj: '', address: '' });
  const [isTwoFactorEnabled, setIsTwoFactorEnabled] = useState(false);
  const [showKeyModal, setShowKeyModal] = useState(false);
  const [generatedApiKey, setGeneratedApiKey] = useState('');
  const [cnpjLoading, setCnpjLoading] = useState(false);
  const [cnpjError, setCnpjError] = useState('');
  const [step1Error, setStep1Error] = useState('');
  const [step2Error, setStep2Error] = useState('');
  const [submitLoading, setSubmitLoading] = useState(false);
  const [submitError, setSubmitError] = useState('');

  const handlePersonalChange = (e) => setPersonalData({ ...personalData, [e.target.name]: e.target.value });
  const handleFactoryChange = (e) => {
    setFactoryData({ ...factoryData, [e.target.name]: e.target.value });
    setCnpjError('');
  };

  const handleBuscarCnpj = async () => {
    const digits = (factoryData.cnpj || '').replace(/\D/g, '');
    if (digits.length !== 14) {
      setCnpjError('Digite um CNPJ válido (14 dígitos) e clique em Buscar.');
      return;
    }
    setCnpjError('');
    setCnpjLoading(true);
    try {
      const res = await api.get(`/api/cnpj?q=${digits}`);
      setFactoryData({
        ...factoryData,
        name: res.data.name || factoryData.name,
        address: res.data.address || factoryData.address,
      });
    } catch (err) {
      setCnpjError(err.response?.status === 404 ? 'CNPJ não encontrado.' : 'Não foi possível buscar os dados. Tente novamente.');
    } finally {
      setCnpjLoading(false);
    }
  };

  const handleSubmit = async () => {
    setSubmitError('');
    setSubmitLoading(true);
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
        const errBody = await response.json().catch(() => ({}));
        throw new Error(errBody.error || errBody.message || 'Falha ao finalizar o cadastro.');
      }

      const data = await response.json();
      setGeneratedApiKey(data.apiKey);
      setShowKeyModal(true);
    } catch (error) {
      setSubmitError(error.message || 'Falha ao finalizar o cadastro. Tente novamente.');
    } finally {
      setSubmitLoading(false);
    }
  };

  const handleCloseModal = () => {
    setShowKeyModal(false);
    // replace evita que o "voltar" do browser retorne ao modal/etapa 3
    window.location.replace('/welcome');
  };

  const handleNextStep = () => {
    if (step === 1) {
      setStep1Error('');
      const name = (personalData.fullName || '').trim();
      const cpf = (personalData.cpf || '').replace(/\D/g, '');
      if (!name) {
        setStep1Error('Preencha o nome completo.');
        return;
      }
      if (cpf.length !== 11) {
        setStep1Error('Preencha um CPF válido (11 dígitos).');
        return;
      }
    }
    if (step === 2) {
      setStep2Error('');
      const name = (factoryData.name || '').trim();
      if (!name) {
        setStep2Error('Preencha o nome da fábrica.');
        return;
      }
    }
    setStep(step + 1);
  };
  const handlePrevStep = () => {
    setStep1Error('');
    setStep2Error('');
    setStep(step - 1);
  };

  return (
    <>
      {showKeyModal && <ApiKeyModal apiKey={generatedApiKey} onClose={handleCloseModal} />}
      <div className={`flex items-center justify-center min-h-screen bg-gray-50 ${showKeyModal ? 'filter blur-sm' : ''}`}>
        <div className="p-8 bg-white rounded-xl shadow-md border border-gray-200 w-full max-w-lg">
          <h1 className="text-2xl font-bold mb-4 text-center">Finalize seu Cadastro</h1>
          {
            {
              1: (
                <div>
                  <h2 className="text-xl font-semibold mb-4">Passo 1: Seus Dados Pessoais</h2>
                  {step1Error && (
                    <div id="onb-step1-error" className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm" role="alert" aria-live="polite">
                      {step1Error}
                    </div>
                  )}
                  <div className="mb-4">
                    <label className="block text-gray-700 font-bold mb-2" htmlFor="onb-fullName">Nome Completo</label>
                    <input
                      id="onb-fullName"
                      type="text"
                      name="fullName"
                      value={personalData.fullName}
                      onChange={(e) => { handlePersonalChange(e); setStep1Error(''); }}
                      className={`w-full px-3 py-2 border rounded-lg ${step1Error ? 'border-red-500' : 'border-gray-300'}`}
                      aria-invalid={!!step1Error}
                      aria-describedby={step1Error ? 'onb-step1-error' : undefined}
                    />
                  </div>
                  <div className="mb-4">
                    <label className="block text-gray-700 font-bold mb-2" htmlFor="onb-cpf">CPF</label>
                    <input
                      id="onb-cpf"
                      type="text"
                      name="cpf"
                      value={personalData.cpf}
                      onChange={(e) => { handlePersonalChange(e); setStep1Error(''); }}
                      placeholder="000.000.000-00"
                      className={`w-full px-3 py-2 border rounded-lg ${step1Error ? 'border-red-500' : 'border-gray-300'}`}
                      aria-invalid={!!step1Error}
                    />
                  </div>
                  <button type="button" onClick={handleNextStep} className="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2 px-4 rounded-lg">
                    Próximo
                  </button>
                </div>
              ),
              2: (
                <div>
                  <h2 className="text-xl font-semibold mb-4">Passo 2: Dados da sua Fábrica</h2>
                  {(step2Error || cnpjError) && (
                    <div id="onb-step2-error" className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm" role="alert" aria-live="polite">
                      {step2Error || cnpjError}
                    </div>
                  )}
                  <div className="mb-4">
                    <label className="block text-gray-700 font-bold mb-2" htmlFor="onb-factory-name">Nome da Fábrica</label>
                    <input
                      id="onb-factory-name"
                      type="text"
                      name="name"
                      value={factoryData.name}
                      onChange={(e) => { handleFactoryChange(e); setStep2Error(''); }}
                      className={`w-full px-3 py-2 border rounded-lg ${step2Error ? 'border-red-500' : 'border-gray-300'}`}
                      aria-invalid={!!step2Error}
                      aria-describedby={step2Error ? 'onb-step2-error' : undefined}
                    />
                  </div>
                  <div className="mb-4">
                    <label className="block text-gray-700 font-bold mb-2">CNPJ</label>
                    <div className="flex gap-2">
                      <input type="text" name="cnpj" value={factoryData.cnpj} onChange={handleFactoryChange} placeholder="00.000.000/0001-00" className="flex-1 px-3 py-2 border rounded-lg border-gray-300" />
                      <button type="button" onClick={handleBuscarCnpj} disabled={cnpjLoading} className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 whitespace-nowrap">
                        {cnpjLoading ? 'Buscando...' : 'Buscar'}
                      </button>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">Busca dados públicos (Receita Federal) para preencher nome e endereço.</p>
                  </div>
                  <div className="mb-4">
                    <label className="block text-gray-700 font-bold mb-2">Endereço</label>
                    <input type="text" name="address" value={factoryData.address} onChange={handleFactoryChange} className="w-full px-3 py-2 border rounded-lg border-gray-300" />
                  </div>
                  <div className="flex justify-between gap-3">
                    <button type="button" onClick={handlePrevStep} className="bg-gray-500 hover:bg-gray-700 text-white font-semibold py-2 px-4 rounded-lg">
                      Anterior
                    </button>
                    <button type="button" onClick={handleNextStep} className="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2 px-4 rounded-lg">
                      Próximo
                    </button>
                  </div>
                </div>
              ),
              3: (
                <div>
                  <h2 className="text-xl font-semibold mb-4">Passo 3: Segurança da Conta</h2>
                  <p className="text-gray-600 mb-4">
                    Opcionalmente você pode ativar autenticação em duas etapas (2FA) depois, em Ajustes. Ao finalizar, sua chave de API será gerada para conectar o DX à plataforma.
                  </p>
                  {submitError && (
                    <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm" role="alert">
                      {submitError}
                    </div>
                  )}
                  <div className="flex justify-between gap-3">
                    <button type="button" onClick={handlePrevStep} className="flex-1 bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded-lg">
                      Anterior
                    </button>
                    <button type="button" onClick={handleSubmit} disabled={submitLoading} className="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-bold py-2 px-4 rounded-lg disabled:opacity-50">
                      {submitLoading ? 'Finalizando…' : 'Finalizar Cadastro e Gerar Chave'}
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
