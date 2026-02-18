import React, { useState } from 'react';
import RegenerateKeyModal from '../components/RegenerateKeyModal';

export default function Settings() {
  const [isModalOpen, setIsModalOpen] = useState(false);

  return (
    <>
      <div className="p-8">
        <h1 className="text-2xl font-bold mb-4">Configurações</h1>
        <div className="bg-white p-6 rounded-lg shadow-md">
          <h2 className="text-xl font-semibold mb-4">Chave de API</h2>
          <p className="text-gray-600 mb-4">
            Ao gerar uma nova chave, sua chave de API atual será invalidada permanentemente. 
            A nova chave gerada só será exibida uma vez por motivos de segurança.
          </p>
          <button
            onClick={() => setIsModalOpen(true)}
            className="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded-lg"
          >
            Gerar Nova Chave de API
          </button>
        </div>
      </div>
      {isModalOpen && <RegenerateKeyModal onClose={() => setIsModalOpen(false)} />}
    </>
  );
}
