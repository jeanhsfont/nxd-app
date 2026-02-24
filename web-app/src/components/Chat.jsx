import React, { useState, useEffect, useRef } from 'react';
import { Send, Bot, User, Cpu, Sparkles } from 'lucide-react';
import api from '../utils/api';

function ThinkingIndicator() {
  return (
    <div className="flex items-start gap-3">
      <div className="w-7 h-7 rounded-full bg-indigo-100 border border-indigo-200 flex items-center justify-center shrink-0 mt-0.5">
        <Bot className="w-3.5 h-3.5 text-indigo-600" />
      </div>
      <div className="bg-white border border-gray-200 rounded-2xl rounded-tl-sm px-4 py-3 shadow-sm">
        <div className="flex items-center gap-1.5">
          <Sparkles className="w-3 h-3 text-indigo-400 animate-pulse mr-1" />
          <span className="text-xs text-gray-400 mr-1">Analisando telemetria</span>
          <span className="w-1.5 h-1.5 bg-indigo-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
          <span className="w-1.5 h-1.5 bg-indigo-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
          <span className="w-1.5 h-1.5 bg-indigo-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
        </div>
      </div>
    </div>
  );
}

function Message({ message }) {
  const isUser = message.sender === 'user';
  return (
    <div className={`flex items-start gap-3 ${isUser ? 'flex-row-reverse' : ''}`}>
      <div className={`w-7 h-7 rounded-full flex items-center justify-center shrink-0 mt-0.5 ${
        isUser ? 'bg-indigo-600' : 'bg-indigo-100 border border-indigo-200'
      }`}>
        {isUser
          ? <User className="w-3.5 h-3.5 text-white" />
          : <Bot className="w-3.5 h-3.5 text-indigo-600" />
        }
      </div>
      <div className={`max-w-[78%] px-4 py-3 rounded-2xl shadow-sm text-sm leading-relaxed whitespace-pre-wrap ${
        isUser
          ? 'bg-indigo-600 text-white rounded-tr-sm'
          : message.error
            ? 'bg-red-50 border border-red-200 text-red-700 rounded-tl-sm'
            : 'bg-white border border-gray-200 text-gray-800 rounded-tl-sm'
      }`}>
        {message.text}
        {message.sectorName && (
          <span className="block mt-1.5 text-xs text-indigo-300 border-t border-indigo-400/20 pt-1">
            📍 Setor: {message.sectorName}
          </span>
        )}
      </div>
    </div>
  );
}

export default function Chat() {
  const [messages, setMessages] = useState([
    {
      sender: 'assistant',
      text: 'Olá! Sou o NXD Intelligence, alimentado pelo Gemini AI.\n\nTenho acesso em tempo real aos dados de telemetria dos seus CLPs. Posso analisar métricas, identificar anomalias, comparar setores e gerar insights operacionais.\n\nComo posso ajudar?',
    }
  ]);
  const [input, setInput] = useState('');
  const [sectors, setSectors] = useState([]);
  const [selectedSector, setSelectedSector] = useState('');
  const [isThinking, setIsThinking] = useState(false);
  const bottomRef = useRef(null);

  useEffect(() => {
    api.get('/api/sectors')
      .then((res) => setSectors(Array.isArray(res.data) ? res.data : []))
      .catch(() => setSectors([]));
  }, []);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isThinking]);

  const handleSend = async () => {
    if (!input.trim() || isThinking) return;

    const sectorObj = sectors.find(s => s.id === selectedSector);
    const userMessage = {
      sender: 'user',
      text: input.trim(),
      sectorName: sectorObj?.name || null,
    };
    setMessages(prev => [...prev, userMessage]);
    const currentInput = input.trim();
    setInput('');
    setIsThinking(true);

    try {
      const res = await api.post('/api/ia/chat', {
        message: currentInput,
        sector_id: selectedSector || undefined,
      });
      setMessages(prev => [...prev, {
        sender: 'assistant',
        text: res.data.reply || 'Sem resposta.',
      }]);
    } catch (err) {
      const errMsg = err?.response?.data || err?.message || 'Erro ao contatar a IA.';
      setMessages(prev => [...prev, {
        sender: 'assistant',
        text: `⚠️ ${errMsg}`,
        error: true,
      }]);
    } finally {
      setIsThinking(false);
    }
  };

  const suggestions = [
    'Quais CLPs estão com problemas agora?',
    'Qual é o Health Score médio da fábrica?',
    'Há alguma anomalia nas últimas leituras?',
    'Compare o consumo de energia dos CLPs',
  ];

  return (
    <div className="flex flex-col h-full min-h-0">
      {/* Messages area */}
      <div className="flex-1 overflow-y-auto px-2 py-4 space-y-4">
        {messages.map((msg, i) => (
          <Message key={i} message={msg} />
        ))}
        {isThinking && <ThinkingIndicator />}
        <div ref={bottomRef} />
      </div>

      {/* Sugestões rápidas (apenas quando sem histórico) */}
      {messages.length === 1 && !isThinking && (
        <div className="px-2 pb-3 flex flex-wrap gap-2">
          {suggestions.map((s, i) => (
            <button
              key={i}
              onClick={() => { setInput(s); }}
              className="text-xs px-3 py-1.5 bg-indigo-50 hover:bg-indigo-100 text-indigo-600 border border-indigo-200 rounded-full transition-colors"
            >
              {s}
            </button>
          ))}
        </div>
      )}

      {/* Input bar */}
      <div className="border-t border-gray-100 pt-4 mt-2">
        {sectors.length > 0 && (
          <div className="flex items-center gap-2 mb-3">
            <Cpu className="w-4 h-4 text-gray-400 shrink-0" />
            <select
              value={selectedSector}
              onChange={(e) => setSelectedSector(e.target.value)}
              className="text-sm text-gray-600 border border-gray-200 rounded-lg px-3 py-1.5 bg-white focus:outline-none focus:ring-2 focus:ring-indigo-300"
            >
              <option value="">🏭 Toda a fábrica</option>
              {sectors.map((s) => (
                <option key={s.id} value={s.id}>📍 {s.name}</option>
              ))}
            </select>
            {selectedSector && (
              <span className="text-xs text-indigo-500 bg-indigo-50 px-2 py-0.5 rounded-full">
                filtro ativo
              </span>
            )}
          </div>
        )}
        <div className="flex items-center gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && handleSend()}
            disabled={isThinking}
            placeholder={isThinking ? 'Consultando Gemini AI...' : 'Pergunte sobre CLPs, métricas, anomalias, setores...'}
            className="flex-1 px-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-300 bg-white disabled:bg-gray-50 disabled:text-gray-400 transition-all"
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || isThinking}
            className="w-10 h-10 bg-indigo-600 hover:bg-indigo-700 disabled:bg-gray-200 text-white rounded-xl flex items-center justify-center transition-all shrink-0"
          >
            <Send className="w-4 h-4" />
          </button>
        </div>
        <p className="text-xs text-gray-400 mt-2 text-center">
          Gemini AI · dados em tempo real dos seus CLPs
        </p>
      </div>
    </div>
  );
}
