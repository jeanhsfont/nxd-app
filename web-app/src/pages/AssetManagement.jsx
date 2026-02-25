import React, { useState, useEffect } from 'react';
import {
  DndContext,
  DragOverlay,
  useSensor,
  useSensors,
  PointerSensor,
  closestCenter,
} from '@dnd-kit/core';
import { useDraggable, useDroppable } from '@dnd-kit/core';
import { CSS } from '@dnd-kit/utilities';
import { Plus, Trash2, GripVertical, Factory, Box, KeyRound, Copy } from 'lucide-react';
import api from '../utils/api';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

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
    <div className="nxd-card flex items-center justify-between">
      <div className="flex items-center gap-3">
        <KeyRound className="w-5 h-5 text-navy" />
        <span className="text-sm text-gray-700 font-mono break-all">{apiKey}</span>
      </div>
      <button
        onClick={copyToClipboard}
        className="nxd-btn nxd-btn-primary"
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
        "flex items-center gap-3 p-3 nxd-card cursor-grab hover:shadow-md transition-all",
        isDragging && "opacity-50 ring-2 ring-navy z-50"
      )}
    >
      <GripVertical className="text-gray-400 w-5 h-5" />
      <div className="flex-1">
        <p className="font-medium text-gray-900">{machine.display_name || machine.name}</p>
        <div className="flex items-center gap-2 text-xs text-gray-500 mt-1">
          <span className={cn(
            "status-dot",
            machine.status === 'online' ? "status-online" : "status-error"
          )} />
          {machine.status}
        </div>
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
    <aside
      ref={setNodeRef}
      className={cn(
        "w-80 bg-white border-r border-gray-200 flex flex-col transition-colors",
        isOver && "bg-navy/5"
      )}
    >
      <div className="p-6 border-b border-gray-200">
        <div className="flex items-center gap-3 mb-2">
          <Box className="w-6 h-6 text-navy" />
          <h2 className="text-lg font-bold text-gray-900">Ativos Disponíveis</h2>
        </div>
        <p className="text-sm text-gray-500">Arraste para os setores →</p>
      </div>
      <div className="flex flex-col gap-3 px-6 pb-6 pt-6 overflow-y-auto">
        {machines.map(m => (
          <DraggableMachine key={m.id} machine={m} />
        ))}
        {machines.length === 0 && (
          <div className="text-center py-10 text-gray-400">
            <Box className="w-12 h-12 mx-auto mb-3 opacity-20" />
            <p className="text-sm">Todos os ativos estão<br/>associados a setores</p>
          </div>
        )}
      </div>
    </aside>
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
        "bg-gray-50 border-2 border-dashed border-gray-300 rounded-lg p-4 min-h-[200px] flex flex-col gap-3 transition-colors",
        isOver && "bg-navy/5 border-navy"
      )}
    >
      <div className="flex justify-between items-center">
        <h3 className="font-semibold text-gray-900">{sector.name}</h3>
        <button
          onClick={() => onDelete(sector.id)}
          className="text-red hover:bg-red/10 p-1 rounded transition-colors"
        >
          <Trash2 className="w-4 h-4" />
        </button>
      </div>
      <div className="flex-1 space-y-2">
        {machines.map(m => (
          <DraggableMachine key={m.id} machine={m} />
        ))}
        {machines.length === 0 && (
          <div className="flex items-center justify-center h-full text-gray-400 text-sm">
            Arraste ativos aqui
          </div>
        )}
      </div>
    </div>
  );
}

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
        const assets = (dashboardRes.data.assets || []).map(a => ({
          id: a.id,
          name: a.display_name,
          display_name: a.display_name,
          source_tag_id: a.source_tag_id,
          status: a.is_online ? 'online' : 'offline',
          sector_id: a.group_id || null,
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
      newSectorId = null;
    } else {
      const targetSector = sectors.find(s => s.id === overId);
      if (targetSector) {
        newSectorId = targetSector.id;
      }
    }

    if (newSectorId !== undefined) {
      const prevMachines = machines;

      setMachines(prev => prev.map(m =>
        m.id === machineId ? { ...m, sector_id: newSectorId } : m
      ));

      try {
        await api.put('/api/machine/asset', {
          machine_id: machineId,
          sector_id: newSectorId === null ? '' : newSectorId,
        });
      } catch (error) {
        console.error("Erro ao atualizar setor da máquina:", error);
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
      await api.delete(`/api/sectors/${id}`);
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
      <div className="flex h-[calc(100vh-64px)] bg-gray-50 font-sans overflow-hidden">
        <Sidebar machines={machines.filter(m => !m.sector_id)} />

        <main className="flex-1 p-6 overflow-y-auto">
          <header className="mb-6">
            <div className="flex justify-between items-center mb-6">
              <div>
                <h1 className="text-3xl font-bold text-gray-900">Gestão de Ativos</h1>
                <p className="text-gray-500 mt-1">Organize suas máquinas em setores para análise de IA</p>
              </div>
              
              <div className="flex gap-2">
                <input
                  type="text"
                  placeholder="Nome do novo setor..."
                  className="nxd-input"
                  value={newSectorName}
                  onChange={e => setNewSectorName(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && createSector()}
                />
                <button
                  onClick={createSector}
                  className="nxd-btn nxd-btn-primary"
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
              <div className="col-span-full flex flex-col items-center justify-center py-20 text-gray-400 border-2 border-dashed border-gray-300 rounded-lg bg-white">
                <Factory className="w-16 h-16 mb-4 opacity-20" />
                <p className="text-lg font-semibold text-gray-600">Nenhum setor criado</p>
                <p className="text-sm text-gray-500">Crie um setor acima para começar a organizar</p>
              </div>
            )}
          </div>
        </main>
      </div>

      <DragOverlay>
        {activeMachine ? (
          <div className="opacity-90 rotate-3 cursor-grabbing pointer-events-none">
             <div className="nxd-card flex items-center gap-3 p-3 w-64 ring-2 ring-navy">
              <GripVertical className="text-navy w-5 h-5" />
              <div>
                <p className="font-medium text-gray-900">{activeMachine.display_name}</p>
                <div className="flex items-center gap-2 text-xs text-gray-500">
                  <span className={cn(
                    "status-dot",
                    activeMachine.status === 'online' ? "status-online" : "status-error"
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
