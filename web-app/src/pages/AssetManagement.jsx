import React, { useState, useEffect } from 'react';
import {
  DndContext,
  DragOverlay,
  useSensor,
  useSensors,
  PointerSensor,
  closestCenter,
} from '@dnd-kit/core';
import {
  SortableContext,
  arrayMove,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { useDraggable, useDroppable } from '@dnd-kit/core';
import { CSS } from '@dnd-kit/utilities';
import { Plus, Trash2, GripVertical, Factory, Box, Server, KeyRound, Copy } from 'lucide-react';
import api from '../utils/api';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

// --- Utils ---
function cn(...inputs) {
  return twMerge(clsx(inputs));
}

// --- Components ---

function APIKeyDisplay({ apiKey }) {
  const [copied, setCopied] = useState(false);

  const copyToClipboard = () => {
    navigator.clipboard.writeText(apiKey).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  if (!apiKey) return null;

  return (
    <div className="bg-gray-800 p-4 rounded-lg flex items-center justify-between">
      <div className="flex items-center gap-3">
        <KeyRound className="w-5 h-5 text-yellow-400" />
        <span className="text-sm text-gray-300 font-mono break-all">{apiKey}</span>
      </div>
      <button
        onClick={copyToClipboard}
        className="bg-gray-700 text-white px-3 py-1 rounded-md hover:bg-gray-600 text-sm flex items-center gap-2"
      >
        <Copy className="w-4 h-4" />
        {copied ? 'Copiado!' : 'Copiar'}
      </button>
    </div>
  );
}
function DraggableMachine({ machine }) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: machine.id,
    data: { machine },
  });

  const style = transform ? {
    transform: CSS.Translate.toString(transform),
  } : undefined;

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      className={cn(
        "flex items-center gap-3 p-3 bg-white border rounded-lg shadow-sm cursor-grab hover:shadow-md transition-shadow",
        isDragging && "opacity-50 ring-2 ring-blue-500 z-50"
      )}
    >
      <GripVertical className="text-gray-400 w-5 h-5" />
      <div>
        <p className="font-medium text-gray-900">{machine.display_name || machine.name}</p>
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <span className={cn(
            "w-2 h-2 rounded-full",
            machine.status === 'online' ? "bg-green-500" : "bg-red-500"
          )} />
          {machine.status}
        </div>
      </div>
    </div>
  );
}

function DroppableSector({ sector, machines, onDelete }) {
  const { setNodeRef, isOver } = useDroppable({
    id: sector.id,
    data: { type: 'sector', sector },
  });

  return (
    <div
      ref={setNodeRef}
      className={cn(
        "bg-gray-50 border-2 border-dashed rounded-xl p-4 min-h-[200px] flex flex-col gap-3 transition-colors",
        isOver ? "border-blue-500 bg-blue-50" : "border-gray-200"
      )}
    >
      <div className="flex items-center justify-between mb-2">
        <h3 className="font-semibold text-gray-700 flex items-center gap-2">
          <Box className="w-4 h-4" />
          {sector.name}
        </h3>
        <button
          onClick={() => onDelete(sector.id)}
          className="text-gray-400 hover:text-red-500 transition-colors"
        >
          <Trash2 className="w-4 h-4" />
        </button>
      </div>
      
      <div className="flex flex-col gap-2 flex-1">
        {machines.map(m => (
          <DraggableMachine key={m.id} machine={m} />
        ))}
        {machines.length === 0 && (
          <div className="flex-1 flex items-center justify-center text-gray-400 text-sm italic">
            Arraste máquinas aqui
          </div>
        )}
      </div>
    </div>
  );
}

function Sidebar({ machines }) {
  const { setNodeRef, isOver } = useDroppable({
    id: 'sidebar-droppable',
    data: { type: 'sidebar' },
  });

  return (
    <div
      ref={setNodeRef}
      className={cn(
        "w-80 bg-white border-r border-gray-200 h-full overflow-y-auto flex flex-col gap-4 flex-shrink-0",
        isOver && "bg-gray-50"
      )}
    >
      <div className="flex items-center gap-2 mb-4 p-6 pb-0">
        <Server className="w-6 h-6 text-indigo-600" />
        <h2 className="text-xl font-bold text-gray-800">Ativos Disponíveis</h2>
      </div>
      
      <div className="flex flex-col gap-3 px-6 pb-6">
        {machines.map(m => (
          <DraggableMachine key={m.id} machine={m} />
        ))}
        {machines.length === 0 && (
          <p className="text-gray-400 text-center py-10">Nenhuma máquina disponível.</p>
        )}
      </div>
    </div>
  );
}

// --- Main App ---

export default function AssetManagement() {
  const [sectors, setSectors] = useState([]);
  const [machines, setMachines] = useState([]);
  const [factory, setFactory] = useState(null);
  const [activeId, setActiveId] = useState(null);
  const [newSectorName, setNewSectorName] = useState('');

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [sectorsRes, dashboardRes, factoryRes] = await Promise.all([
          api.get('/api/sectors'),
          api.get('/api/dashboard/data'),
          api.get('/api/factory/details')
        ]);
        setSectors(Array.isArray(sectorsRes.data) ? sectorsRes.data : (sectorsRes.data.sectors || []));
        // Mapeia os assets do dashboard para o formato esperado pelo DnD
        const assets = (dashboardRes.data.assets || []).map(a => ({
          id: a.id,
          name: a.display_name,
          display_name: a.display_name,
          source_tag_id: a.source_tag_id,
          status: a.is_online ? 'online' : 'offline',
          sector_id: a.group_id || null, // group_id vem do backend (UUID string ou null)
        }));
        setMachines(assets);
        setFactory(factoryRes.data);
      } catch (error) {
        console.error("Erro ao buscar dados:", error);
      }
    };
    fetchData();
  }, []);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  );

  const handleDragStart = (event) => {
    setActiveId(event.active.id);
  };

  const handleDragEnd = async (event) => {
    const { active, over } = event;
    setActiveId(null);

    if (!over) return;

    const machineId = active.id;
    const overId = over.id;
    let newSectorId = undefined;

    if (overId === 'sidebar-droppable') {
      newSectorId = null; // Remove do setor
    } else {
      const targetSector = sectors.find(s => s.id === overId);
      if (targetSector) {
        newSectorId = targetSector.id;
      }
    }

    if (newSectorId !== undefined) {
      // Snapshot anterior para reverter se necessário
      const prevMachines = machines;

      // Atualiza o estado localmente para feedback imediato
      setMachines(prev => prev.map(m =>
        m.id === machineId ? { ...m, sector_id: newSectorId } : m
      ));

      // Chama a API para persistir a mudança (PUT /api/machine/asset)
      try {
        await api.put('/api/machine/asset', {
          machine_id: machineId,
          sector_id: newSectorId === null ? '' : newSectorId,
        });
      } catch (error) {
        console.error("Erro ao atualizar setor da máquina:", error);
        // Reverte o estado local se a API falhar
        setMachines(prevMachines);
      }
    }
  };

  const createSector = async () => {
    if (!newSectorName.trim()) return;
    try {
      const res = await api.post('/api/sectors', { name: newSectorName });
      setSectors([...sectors, res.data]);
      setNewSectorName('');
    } catch (error) {
      console.error("Erro ao criar setor:", error);
    }
  };

  const deleteSector = async (id) => {
    try {
      await api.delete(`/api/sectors/${id}`); // Assumindo que a API suporta DELETE
      setSectors(prev => prev.filter(s => s.id !== id));
      setMachines(prev => prev.map(m =>
        m.sector_id === id ? { ...m, sector_id: null } : m
      ));
    } catch (error) {
      console.error("Erro ao deletar setor:", error);
    }
  };

  const activeMachine = activeId ? machines.find(m => m.id === activeId) : null;

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="flex h-[calc(100vh-64px)] bg-gray-100 font-sans overflow-hidden">
        {/* Sidebar */}
        <Sidebar machines={machines.filter(m => !m.sector_id)} />

        {/* Main Content */}
        <main className="flex-1 p-8 overflow-y-auto">
          <header className="mb-8">
            <div className="flex justify-between items-center mb-4">
              <div>
                <h1 className="text-3xl font-bold text-gray-900">Gestão de Ativos</h1>
                <p className="text-gray-500 mt-1">Organize suas máquinas em setores para análise de IA</p>
              </div>
              
              <div className="flex gap-2">
                <input
                  type="text"
                  placeholder="Nome do novo setor..."
                  className="px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
                  value={newSectorName}
                  onChange={e => setNewSectorName(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && createSector()}
                />
                <button
                  onClick={createSector}
                  className="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 flex items-center gap-2 font-medium transition-colors"
                >
                  <Plus className="w-5 h-5" />
                  Criar Setor
                </button>
              </div>
            </div>
            <APIKeyDisplay apiKey={factory?.api_key} />
          </header>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {sectors.map(sector => (
              <DroppableSector
                key={sector.id}
                sector={sector}
                machines={machines.filter(m => m.sector_id === sector.id)}
                onDelete={deleteSector}
              />
            ))}
            
            {sectors.length === 0 && (
              <div className="col-span-full flex flex-col items-center justify-center py-20 text-gray-400 border-2 border-dashed border-gray-300 rounded-xl bg-gray-50">
                <Factory className="w-16 h-16 mb-4 opacity-20" />
                <p className="text-lg">Nenhum setor criado.</p>
                <p className="text-sm">Crie um setor acima para começar a organizar.</p>
              </div>
            )}
          </div>
        </main>
      </div>

      <DragOverlay>
        {activeMachine ? (
          <div className="opacity-90 rotate-3 cursor-grabbing pointer-events-none">
             <div className="flex items-center gap-3 p-3 bg-white border border-indigo-200 rounded-lg shadow-xl ring-2 ring-indigo-500 w-64">
              <GripVertical className="text-indigo-500 w-5 h-5" />
              <div>
                <p className="font-medium text-gray-900">{activeMachine.display_name}</p>
                <div className="flex items-center gap-2 text-xs text-gray-500">
                  <span className={cn(
                    "w-2 h-2 rounded-full",
                    activeMachine.status === 'online' ? "bg-green-500" : "bg-red-500"
                  )} />
                  {activeMachine.status}
                </div>
              </div>
            </div>
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
