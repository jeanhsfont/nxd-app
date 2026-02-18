import React, { useState, useEffect } from 'react';

function EditSectorModal({ isOpen, onClose, onSave, sector }) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');

  useEffect(() => {
    if (sector) {
      setName(sector.name);
      setDescription(sector.description || '');
    }
  }, [sector]);

  const handleSave = () => {
    onSave({ ...sector, name, description });
  };

  if (!isOpen) {
    return null;
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex justify-center items-center">
      <div className="bg-white p-6 rounded-lg shadow-xl max-w-lg w-full">
        <h3 className="text-lg font-bold mb-4">Editar Setor</h3>
        <div className="space-y-4">
          <div>
            <label htmlFor="editSectorName" className="block text-sm font-medium text-gray-700">
              Nome do Setor
            </label>
            <input
              type="text"
              id="editSectorName"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              required
            />
          </div>
          <div>
            <label htmlFor="editSectorDescription" className="block text-sm font-medium text-gray-700">
              Descrição (Opcional)
            </label>
            <textarea
              id="editSectorDescription"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows="3"
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            ></textarea>
          </div>
        </div>
        <div className="flex justify-end gap-4 mt-6">
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-md text-gray-700 bg-gray-200 hover:bg-gray-300"
          >
            Cancelar
          </button>
          <button
            onClick={handleSave}
            className="px-4 py-2 rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
          >
            Salvar Alterações
          </button>
        </div>
      </div>
    </div>
  );
}

export default EditSectorModal;
