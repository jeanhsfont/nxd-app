import React, { useState, useEffect } from 'react';
import api from '../utils/api';

export default function Chat() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [sectors, setSectors] = useState([]);
  const [selectedSector, setSelectedSector] = useState('');

  useEffect(() => {
    // TODO: Replace with actual factory_id
    const factoryId = 'f47ac10b-58cc-4372-a567-0e02b2c3d479';
    api.get(`/api/groups?factory_id=${factoryId}`)
      .then((res) => {
        setSectors(res.data.groups);
      });
  }, []);

  const handleSend = () => {
    if (input.trim()) {
      const message = {
        text: input,
        sender: 'user',
        sector: selectedSector,
      };
      setMessages([...messages, message]);
      setInput('');
      console.log('Sending message:', message);
      // TODO: Add logic to send message to AI and receive response
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 overflow-y-auto p-4">
        {messages.map((message, index) => (
          <div
            key={index}
            className={`p-2 my-2 rounded-lg ${
              message.sender === 'user' ? 'bg-blue-500 text-white self-end' : 'bg-gray-300 text-black self-start'
            }`}
          >
            {message.text}
          </div>
        ))}
      </div>
      <div className="p-4 flex items-center">
        <select
          value={selectedSector}
          onChange={(e) => setSelectedSector(e.target.value)}
          className="p-2 border rounded-lg mr-2"
        >
          <option value="">Todos os setores</option>
          {sectors.map((sector) => (
            <option key={sector.id} value={sector.id}>
              {sector.name}
            </option>
          ))}
        </select>
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyPress={(e) => e.key === 'Enter' && handleSend()}
          className="flex-1 p-2 border rounded-lg"
          placeholder="Digite sua mensagem..."
        />
        <button onClick={handleSend} className="ml-2 p-2 bg-blue-500 text-white rounded-lg">
          Enviar
        </button>
      </div>
    </div>
  );
}
