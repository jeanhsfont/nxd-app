import React, { useState, useEffect } from 'react';
import api from '../utils/api';
import { Plus, Trash2, Edit, Package, Loader2 } from 'lucide-react';
import toast from 'react-hot-toast';
import ConfirmationModal from '../components/ConfirmationModal';
import EditSectorModal from '../components/EditSectorModal';

function Sectors() {
  const [sectors, setSectors] = useState([]);
  const [loading, setLoading] = useState(true);
  const [newSectorName, setNewSectorName] = useState('');
  const [newSectorDescription, setNewSectorDescription] = useState('');
  const [creating, setCreating] = useState(false);
  
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [sectorToDelete, setSectorToDelete] = useState(null);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [sectorToEdit, setSectorToEdit] = useState(null);

  useEffect(() => {
    fetchSectors();
  }, []);

  const fetchSectors = async () => {
    setLoading(true);
    try {
      const response = await api.get('/api/sectors');
      setSectors(Array.isArray(response.data) ? response.data : (response.data?.sectors || []));
    } catch (err) {
      toast.error('Não foi possível carregar os setores');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateSector = async (e) => {
    e.preventDefault();
    if (!newSectorName.trim()) {
      toast.error('O nome do setor é obrigatório');
      return;
    }

    setCreating(true);
    try {
      const response = await api.post('/api/sectors', {
        name: newSectorName,
        description: newSectorDescription,
      });
      setSectors([...sectors, response.data]);
      setNewSectorName('');
      setNewSectorDescription('');
      toast.success('Setor criado com sucesso!');
    } catch (err) {
      toast.error('Não foi possível criar o setor');
    } finally {
      setCreating(false);
    }
  };

  const openDeleteModal = (sector) => {
    setSectorToDelete(sector);
    setIsDeleteModalOpen(true);
  };

  const handleDeleteSector = async () => {
    if (!sectorToDelete) return;

    try {
      await api.delete(`/api/sectors/${sectorToDelete.id}`);
      setSectors(sectors.filter((s) => s.id !== sectorToDelete.id));
      toast.success('Setor excluído!');
      setIsDeleteModalOpen(false);
    } catch (err) {
      toast.error('Não foi possível excluir o setor');
    }
  };

  const handleUpdateSector = async (updatedSector) => {
    try {
      const response = await api.put(`/api/sectors/${updatedSector.id}`, updatedSector);
      setSectors(sectors.map((s) => (s.id === updatedSector.id ? response.data : s)));
      setIsEditModalOpen(false);
      toast.success('Setor atualizado!');
    } catch (err) {
      toast.error('Não foi possível atualizar o setor');
    }
  };

  return (
    <>
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-6xl mx-auto p-6">
          {/* Header */}
          <div className="page-header">
            <div className="page-header-icon">
              <Package className="w-6 h-6" />
            </div>
            <div>
              <h1 className="page-title">Gestão de Setores</h1>
              <p className="page-subtitle">Organize suas máquinas por setores</p>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Create form */}
            <div className="lg:col-span-1">
              <div className="nxd-card sticky top-6">
                <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                  <Plus className="w-5 h-5 text-green" />
                  Criar Novo Setor
                </h2>
                <form onSubmit={handleCreateSector} className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Nome do Setor
                    </label>
                    <input
                      type="text"
                      value={newSectorName}
                      onChange={(e) => setNewSectorName(e.target.value)}
                      className="nxd-input"
                      placeholder="Ex: Usinagem"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Descrição (Opcional)
                    </label>
                    <textarea
                      value={newSectorDescription}
                      onChange={(e) => setNewSectorDescription(e.target.value)}
                      rows="3"
                      className="nxd-input resize-none"
                      placeholder="Breve descrição do setor"
                    />
                  </div>
                  <button
                    type="submit"
                    disabled={creating}
                    className="nxd-btn nxd-btn-primary w-full justify-center"
                  >
                    {creating ? (
                      <>
                        <Loader2 className="w-5 h-5 animate-spin" />
                        Criando...
                      </>
                    ) : (
                      <>
                        <Plus className="w-5 h-5" />
                        Criar Setor
                      </>
                    )}
                  </button>
                </form>
              </div>
            </div>

            {/* Sectors List */}
            <div className="lg:col-span-2">
              {loading ? (
                <div className="text-center py-12">
                  <div className="spinner mx-auto mb-4"></div>
                  <p className="text-gray-600">Carregando setores...</p>
                </div>
              ) : sectors.length === 0 ? (
                <div className="nxd-card text-center py-12">
                  <Package className="w-16 h-16 text-gray-300 mx-auto mb-4" />
                  <h3 className="text-xl font-bold text-gray-900 mb-2">Nenhum setor cadastrado</h3>
                  <p className="text-gray-600">Crie seu primeiro setor para organizar os ativos</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {sectors.map((sector) => (
                    <div key={sector.id} className="nxd-card fade-in">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h3 className="text-lg font-bold text-gray-900 mb-1">{sector.name}</h3>
                          {sector.description && (
                            <p className="text-gray-600 text-sm">{sector.description}</p>
                          )}
                        </div>
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => {
                              setSectorToEdit(sector);
                              setIsEditModalOpen(true);
                            }}
                            className="p-2 text-navy hover:bg-navy/10 rounded-lg transition-colors"
                            title="Editar"
                          >
                            <Edit className="w-5 h-5" />
                          </button>
                          <button
                            onClick={() => openDeleteModal(sector)}
                            className="p-2 text-red hover:bg-red/10 rounded-lg transition-colors"
                            title="Excluir"
                          >
                            <Trash2 className="w-5 h-5" />
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Modals */}
      {isDeleteModalOpen && (
        <ConfirmationModal
          title="Excluir Setor"
          message={`Tem certeza que deseja excluir o setor "${sectorToDelete?.name}"?`}
          onConfirm={handleDeleteSector}
          onCancel={() => setIsDeleteModalOpen(false)}
        />
      )}
      {isEditModalOpen && sectorToEdit && (
        <EditSectorModal
          sector={sectorToEdit}
          onSave={handleUpdateSector}
          onClose={() => setIsEditModalOpen(false)}
        />
      )}
    </>
  );
}

export default Sectors;
