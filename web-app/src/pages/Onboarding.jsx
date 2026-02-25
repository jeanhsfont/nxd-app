import React, { useState } from 'react';
import { User, Building2, Shield, ArrowRight, ArrowLeft, CheckCircle, Search, Loader2 } from 'lucide-react';
import toast from 'react-hot-toast';
import ApiKeyModal from '../components/ApiKeyModal';
import api from '../utils/api';

const steps = [
  { number: 1, title: 'Dados Pessoais', icon: User },
  { number: 2, title: 'Sua Fábrica', icon: Building2 },
  { number: 3, title: 'Segurança', icon: Shield },
];

export default function Onboarding() {
  const [step, setStep] = useState(1);
  const [personalData, setPersonalData] = useState({ fullName: '', cpf: '' });
  const [factoryData, setFactoryData] = useState({ name: '', cnpj: '', address: '' });
  const [showKeyModal, setShowKeyModal] = useState(false);
  const [generatedApiKey, setGeneratedApiKey] = useState('');
  const [cnpjLoading, setCnpjLoading] = useState(false);
  const [submitLoading, setSubmitLoading] = useState(false);

  const handlePersonalChange = (e) => setPersonalData({ ...personalData, [e.target.name]: e.target.value });
  
  const handleFactoryChange = (e) => {
    setFactoryData({ ...factoryData, [e.target.name]: e.target.value });
  };

  const handleBuscarCnpj = async () => {
    const digits = (factoryData.cnpj || '').replace(/\D/g, '');
    if (digits.length !== 14) {
      toast.error('Digite um CNPJ válido (14 dígitos)');
      return;
    }
    
    setCnpjLoading(true);
    try {
      const res = await api.get(`/api/cnpj?q=${digits}`);
      setFactoryData({
        ...factoryData,
        name: res.data.name || factoryData.name,
        address: res.data.address || factoryData.address,
      });
      toast.success('Dados encontrados!');
    } catch (err) {
      toast.error(err.response?.status === 404 ? 'CNPJ não encontrado' : 'Erro ao buscar dados');
    } finally {
      setCnpjLoading(false);
    }
  };

  const handleSubmit = async () => {
    setSubmitLoading(true);
    try {
      const response = await api.post('/api/onboarding', {
        personalData,
        factoryData,
        twoFactorEnabled: false,
      });
      const data = response.data;
      setGeneratedApiKey(data.apiKey || data.api_key);
      setShowKeyModal(true);
      toast.success('Cadastro finalizado!');
    } catch (error) {
      const msg = error.response?.data?.error || error.message;
      toast.error(msg);
    } finally {
      setSubmitLoading(false);
    }
  };

  const handleCloseModal = () => {
    setShowKeyModal(false);
    window.location.replace('/welcome');
  };

  const handleNextStep = () => {
    if (step === 1) {
      if (!personalData.fullName?.trim()) {
        toast.error('Preencha o nome completo');
        return;
      }
      if ((personalData.cpf || '').replace(/\D/g, '').length !== 11) {
        toast.error('Preencha um CPF válido');
        return;
      }
    }
    if (step === 2) {
      if (!factoryData.name?.trim()) {
        toast.error('Preencha o nome da fábrica');
        return;
      }
      if ((factoryData.cnpj || '').replace(/\D/g, '').length !== 14) {
        toast.error('Preencha um CNPJ válido');
        return;
      }
      if (!factoryData.address?.trim()) {
        toast.error('Preencha o endereço');
        return;
      }
    }
    setStep(step + 1);
  };

  return (
    <>
      {showKeyModal && <ApiKeyModal apiKey={generatedApiKey} onClose={handleCloseModal} />}
      
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="w-full max-w-2xl">
          {/* Progress steps */}
          <div className="mb-8">
            <div className="flex items-center justify-between">
              {steps.map((s, i) => (
                <React.Fragment key={s.number}>
                  <div className={`flex items-center gap-3 flex-1 ${i < steps.length - 1 ? 'pr-4' : ''}`}>
                    <div className={`w-12 h-12 rounded-lg flex items-center justify-center font-bold transition-all ${
                      step >= s.number
                        ? 'bg-navy text-white'
                        : 'bg-gray-200 text-gray-400'
                    }`}>
                      {step > s.number ? <CheckCircle className="w-6 h-6" /> : <s.icon className="w-6 h-6" />}
                    </div>
                    <div className="hidden md:block">
                      <p className={`text-xs ${step >= s.number ? 'text-gray-600' : 'text-gray-400'}`}>
                        Passo {s.number}
                      </p>
                      <p className={`text-sm font-medium ${step >= s.number ? 'text-gray-900' : 'text-gray-500'}`}>
                        {s.title}
                      </p>
                    </div>
                  </div>
                  {i < steps.length - 1 && (
                    <div className={`h-0.5 w-full max-w-[100px] transition-all ${
                      step > s.number ? 'bg-navy' : 'bg-gray-200'
                    }`} />
                  )}
                </React.Fragment>
              ))}
            </div>
          </div>

          {/* Card */}
          <div className="nxd-card fade-in">
            {/* Step 1 */}
            {step === 1 && (
              <div className="space-y-5">
                <h2 className="text-2xl font-bold text-gray-900 mb-6">Seus Dados Pessoais</h2>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Nome Completo</label>
                  <input
                    type="text"
                    name="fullName"
                    value={personalData.fullName}
                    onChange={handlePersonalChange}
                    className="nxd-input"
                    placeholder="João Silva"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">CPF</label>
                  <input
                    type="text"
                    name="cpf"
                    value={personalData.cpf}
                    onChange={handlePersonalChange}
                    className="nxd-input"
                    placeholder="000.000.000-00"
                  />
                </div>
              </div>
            )}

            {/* Step 2 */}
            {step === 2 && (
              <div className="space-y-5">
                <h2 className="text-2xl font-bold text-gray-900 mb-6">Dados da sua Fábrica</h2>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Nome da Fábrica</label>
                  <input
                    type="text"
                    name="name"
                    value={factoryData.name}
                    onChange={handleFactoryChange}
                    className="nxd-input"
                    placeholder="Indústria XYZ"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">CNPJ</label>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      name="cnpj"
                      value={factoryData.cnpj}
                      onChange={handleFactoryChange}
                      className="nxd-input flex-1"
                      placeholder="00.000.000/0001-00"
                    />
                    <button
                      type="button"
                      onClick={handleBuscarCnpj}
                      disabled={cnpjLoading}
                      className="nxd-btn nxd-btn-primary"
                    >
                      {cnpjLoading ? (
                        <Loader2 className="w-5 h-5 animate-spin" />
                      ) : (
                        <Search className="w-5 h-5" />
                      )}
                    </button>
                  </div>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Endereço Completo</label>
                  <textarea
                    name="address"
                    value={factoryData.address}
                    onChange={handleFactoryChange}
                    className="nxd-input"
                    rows="3"
                    placeholder="Rua, Número, Bairro, Cidade - Estado"
                  />
                </div>
              </div>
            )}

            {/* Step 3 */}
            {step === 3 && (
              <div className="space-y-6">
                <h2 className="text-2xl font-bold text-gray-900 mb-6">Segurança da Conta</h2>
                
                <div className="nxd-card bg-gray-50">
                  <div className="flex items-start gap-4">
                    <Shield className="w-8 h-8 text-navy flex-shrink-0" />
                    <div>
                      <h3 className="font-semibold text-gray-900 mb-2">Autenticação de Dois Fatores (2FA)</h3>
                      <p className="text-sm text-gray-600 mb-4">
                        A autenticação de dois fatores adiciona uma camada extra de segurança à sua conta.
                        Você poderá ativar esse recurso mais tarde nas configurações.
                      </p>
                      <p className="text-xs text-gray-500">
                        Por padrão, seu cadastro será criado com segurança básica. Recomendamos ativar o 2FA após o login.
                      </p>
                    </div>
                  </div>
                </div>

                <div className="bg-navy/5 border-l-4 border-navy rounded-lg p-4">
                  <p className="text-sm text-gray-700">
                    <strong>Resumo do Cadastro:</strong>
                  </p>
                  <ul className="mt-2 space-y-1 text-sm text-gray-600">
                    <li>• Nome: {personalData.fullName}</li>
                    <li>• Fábrica: {factoryData.name}</li>
                    <li>• CNPJ: {factoryData.cnpj}</li>
                  </ul>
                </div>
              </div>
            )}

            {/* Navigation Buttons */}
            <div className="flex items-center justify-between mt-8 pt-6 border-t border-gray-200">
              {step > 1 ? (
                <button
                  onClick={() => setStep(step - 1)}
                  className="flex items-center gap-2 text-gray-600 hover:text-navy font-medium transition-colors"
                >
                  <ArrowLeft className="w-4 h-4" />
                  Voltar
                </button>
              ) : <div></div>}

              {step < 3 ? (
                <button
                  onClick={handleNextStep}
                  className="nxd-btn nxd-btn-primary ml-auto"
                >
                  Próximo
                  <ArrowRight className="w-5 h-5" />
                </button>
              ) : (
                <button
                  onClick={handleSubmit}
                  disabled={submitLoading}
                  className="nxd-btn nxd-btn-primary ml-auto"
                >
                  {submitLoading ? (
                    <>
                      <Loader2 className="w-5 h-5 animate-spin" />
                      Finalizando...
                    </>
                  ) : (
                    <>
                      Finalizar Cadastro
                      <CheckCircle className="w-5 h-5" />
                    </>
                  )}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
