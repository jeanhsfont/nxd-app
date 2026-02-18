import React, { useState, useEffect } from 'react';
import api from '../utils/api';
import { Plus, Trash2, Edit } from 'lucide-react';
import ConfirmationModal from '../components/ConfirmationModal';
import EditSectorModal from '../components/EditSectorModal';

function Sectors() {
  const [sectors, setSectors] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [newSectorName, setNewSectorName] = useState('');
  const [newSectorDescription, setNewSectorDescription] = useState('');
  
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [sectorToDelete, setSectorToDelete] = useState(null);

  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [sectorToEdit, setSectorToEdit] = useState(null);

  useEffect(() => {
    const fetchSectors = async () => {
      try {
        const response = await api.get('/api/sectors');
        setSectors(response.data.sectors || []);
      } catch (err) {
        setError('Não foi possível carregar os setores.');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchSectors();
  }, []);

  const handleCreateSector = async (e) => {
    e.preventDefault();
    if (!newSectorName.trim()) {
      alert('O nome do setor é obrigatório.');
      return;
    }

    try {
      const response = await api.post('/api/sectors', {
        name: newSectorName,
        description: newSectorDescription,
      });
      setSectors([...sectors, response.data]);
      setNewSectorName('');
      setNewSectorDescription('');
    } catch (err) {
      setError('Não foi possível criar o setor.');
      console.error(err);
    }
  };

  const openDeleteModal = (sector) => {
    setSectorToDelete(sector);
    setIsDeleteModalOpen(true);
  };

  const closeDeleteModal = () => {
    setSectorToDelete(null);
    setIsDeleteModalOpen(false);
  };

  const handleDeleteSector = async () => {
    if (!sectorToDelete) return;

    try {
      await api.delete(`/api/sectors/${sectorToDelete.id}`);
      setSectors(sectors.filter((s) => s.id !== sectorToDelete.id));
      closeDeleteModal();
    } catch (err) {
      setError('Não foi possível excluir o setor.');
      console.error(err);
      closeDeleteModal();
    }
  };

  const openEditModal = (sector) => {
    setSectorToEdit(sector);
    setIsEditModalOpen(true);
  };

  const closeEditModal = () => {
    setSectorToEdit(null);
    setIsEditModalOpen(false);
  };

  const handleUpdateSector = async (updatedSector) => {
    try {
      const response = await api.put(`/api/sectors/${updatedSector.id}`, updatedSector);
      setSectors(
        sectors.map((s) => (s.id === updatedSector.id ? response.data : s))
      );
      closeEditModal();
    } catch (err) {
      setError('Não foi possível atualizar o setor.');
      console.error(err);
      closeEditModal();
    }
  };

  if (loading) {
    return <div className="container mx-auto p-4">Carregando...</div>;
  }

  if (error) {
    return <div className="container mx-auto p-4 text-red-500">{error}</div>;
  }

  return (
    <>
      <div className="container mx-auto p-8">
        <h1 className="text-3xl font-bold mb-6">Gestão de Setores</h1>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {/* Formulário de Criação */}
          <div className="md:col-span-1">
            <div className="bg-white p-6 rounded-lg shadow-md">
              <h2 className="text-xl font-semibold mb-4">Criar Novo Setor</h2>
              <form onSubmit={handleCreateSector} className="space-y-4">
                <div>
                  <label htmlFor="sectorName" className="block text-sm font-medium text-gray-700">
                    Nome do Setor
                  </label>
                  <input
                    type="text"
                    id="sectorName"
                    value={newSectorName}
                    onChange={(e) => setNewSectorName(e.target.value)}
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                    placeholder="Ex: Usinagem"
                    required
                  />
                </div>
                <div>
                  <label htmlFor="sectorDescription" className="block text-sm font-medium text-gray-700">
                    Descrição (Opcional)
                  </label>
                  <textarea
                    id="sectorDescription"
                    value={newSectorDescription}
                    onChange={(e) => setNewSectorDescription(e.target.value)}
                    rows="3"
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                    placeholder="Descreva brevemente a função deste setor"
                  ></textarea>
                </div>
                <button
                  type="submit"
                  className="w-full flex justify-center items-center gap-2 px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  <Plus className="w-5 h-5" />
                  Criar Setor
                </button>
              </form>
            </div>
          </div>

          {/* Lista de Setores */}
          <div className="md:col-span-2">
            <div className="bg-white p-6 rounded-lg shadow-md">
              <h2 className="text-xl font-semibold mb-4">Setores Existentes</h2>
              {sectors.length === 0 ? (
                <p className="text-gray-500">Nenhum setor criado ainda.</p>
              ) : (
                <ul className="space-y-3">
                  {sectors.map((sector) => (
                    <li
                      key={sector.id}
                      className="flex items-center justify-between bg-gray-50 p-4 rounded-md border"
                    >
                      <div>
                        <p className="font-semibold text-gray-800">{sector.name}</p>
                        {sector.description && (
                          <p className="text-sm text-gray-600">{sector.description}</p>
                        )}
                      </div>
                      <div className="flex items-center gap-3">
                        <button 
                          onClick={() => openEditModal(sector)}
                          className="text-gray-400 hover:text-indigo-600"
                        >
                          <Edit className="w-5 h-5" />
                        </button>
                        <button
                          onClick={() => openDeleteModal(sector)}
                          className="text-gray-400 hover:text-red-600"
                        >
                          <Trash2 className="w-5 h-5" />
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>
        </div>
      </div>
      <ConfirmationModal
        isOpen={isDeleteModalOpen}
        onClose={closeDeleteModal}
        onConfirm={handleDeleteSector}
        title="Confirmar Exclusão"
        message={`Tem certeza que deseja excluir o setor "${sectorToDelete?.name}"? Esta ação não pode ser desfeita.`}
      />
      <EditSectorModal
        isOpen={isEditModalOpen}
        onClose={closeEditModal}
        onSave={handleUpdateSector}
        sector={sectorToEdit}
      />
    </>
  );
}

export default Sectors;
