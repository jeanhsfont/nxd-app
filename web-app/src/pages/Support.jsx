import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { HelpCircle, Send, Mail, LogIn } from 'lucide-react';
import api from '../utils/api';

const SUPPORT_EMAIL = import.meta.env.VITE_SUPPORT_EMAIL || 'suporte@nxd.io';

export default function Support() {
  const [isAuth, setIsAuth] = useState(true);
  const [sent, setSent] = useState(false);
  const [subject, setSubject] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    setIsAuth(!!localStorage.getItem('nxd-token'));
  }, []);

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!subject.trim() || !message.trim()) {
      setError('Preencha assunto e mensagem.');
      return;
    }
    setError('');
    setLoading(true);
    api.post('/api/support', { subject: subject.trim(), message: message.trim() })
      .then(() => setSent(true))
      .catch((err) => setError(err.response?.data?.message || 'Erro ao enviar. Tente novamente.'))
      .finally(() => setLoading(false));
  };

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <div className="flex items-center gap-2 mb-6">
        <HelpCircle className="w-8 h-8 text-indigo-600" />
        <h1 className="text-2xl font-bold text-gray-900">Suporte</h1>
      </div>

      <div className="grid gap-6">
        <div className="bg-white border border-gray-200 rounded-xl shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-800 mb-2">Contato</h2>
          <p className="text-sm text-gray-600 mb-3">
            Envie sua mensagem pelo formulário abaixo ou escreva diretamente para:
          </p>
          <a
            href={`mailto:${SUPPORT_EMAIL}`}
            className="inline-flex items-center gap-2 text-indigo-600 hover:underline"
          >
            <Mail className="w-5 h-5" />
            {SUPPORT_EMAIL}
          </a>
        </div>

        <div className="bg-white border border-gray-200 rounded-xl shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-800 mb-4">Enviar mensagem</h2>
          {!isAuth ? (
            <div className="p-4 bg-amber-50 border border-amber-200 rounded-lg text-amber-800 text-sm flex flex-col gap-3">
              <p>É necessário estar logado para enviar mensagem.</p>
              <Link
                to="/login"
                className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 w-fit font-medium"
              >
                <LogIn className="w-4 h-4" />
                Ir para Login
              </Link>
            </div>
          ) : sent ? (
            <div className="p-4 bg-emerald-50 border border-emerald-200 rounded-lg text-emerald-800 text-sm">
              Mensagem enviada com sucesso. Entraremos em contato em breve.
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && (
                <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
                  {error}
                </div>
              )}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Assunto</label>
                <input
                  type="text"
                  value={subject}
                  onChange={(e) => setSubject(e.target.value)}
                  placeholder="Ex: Dúvida sobre API Key"
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Mensagem</label>
                <textarea
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  rows={4}
                  placeholder="Descreva sua dúvida ou problema..."
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                />
              </div>
              <button
                type="submit"
                disabled={loading}
                className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50"
              >
                <Send className="w-4 h-4" />
                {loading ? 'Enviando...' : 'Enviar'}
              </button>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
