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
import { Plus, Trash2, GripVertical, Factory, Box, Server } from 'lucide-react';
import axios from 'axios';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

// --- Utils ---
function cn(...inputs) {
  return twMerge(clsx(inputs));
}

// --- Components ---

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
  const [activeId, setActiveId] = useState(null);
  const [newSectorName, setNewSectorName] = useState('');

  // Mock initial data load
  useEffect(() => {
    // In real app, fetch from API
    // axios.get('/api/groups').then(...)
    // axios.get('/api/assets').then(...)
    
    setSectors([
      { id: 'sec-1', name: 'Usinagem A', metadata: { color: 'blue' } },
      { id: 'sec-2', name: 'Montagem Final', metadata: { color: 'green' } },
    ]);

    setMachines([
      { id: 'm-1', name: 'CNC-01', display_name: 'CNC Alpha', group_id: 'sec-1', status: 'online' },
      { id: 'm-2', name: 'CNC-02', display_name: 'CNC Beta', group_id: null, status: 'offline' },
      { id: 'm-3', name: 'Robot-Arm', display_name: 'Braço Robótico', group_id: 'sec-2', status: 'online' },
      { id: 'm-4', name: 'Press-Hyd', display_name: 'Prensa Hidráulica', group_id: null, status: 'error' },
    ]);
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

  const handleDragEnd = (event) => {
    const { active, over } = event;
    setActiveId(null);

    if (!over) return;

    const machineId = active.id;
    const overId = over.id;

    // Se soltou na Sidebar (remover do grupo)
    if (overId === 'sidebar-droppable') {
      setMachines(prev => prev.map(m => 
        m.id === machineId ? { ...m, group_id: null } : m
      ));
      // TODO: Call API to update machine group_id = null
      return;
    }

    // Se soltou em um Setor
    // Check if overId is a sector id
    const targetSector = sectors.find(s => s.id === overId);
    if (targetSector) {
      setMachines(prev => prev.map(m => 
        m.id === machineId ? { ...m, group_id: targetSector.id } : m
      ));
      // TODO: Call API to update machine group_id = targetSector.id
    }
  };

  const createSector = () => {
    if (!newSectorName.trim()) return;
    const newSector = {
      id: `sec-${Date.now()}`,
      name: newSectorName,
    };
    setSectors([...sectors, newSector]);
    setNewSectorName('');
    // TODO: Call API to create group
  };

  const deleteSector = (id) => {
    setSectors(prev => prev.filter(s => s.id !== id));
    // Move machines back to available
    setMachines(prev => prev.map(m => 
      m.group_id === id ? { ...m, group_id: null } : m
    ));
    // TODO: Call API to delete group
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
        <Sidebar machines={machines.filter(m => m.group_id === null)} />

        {/* Main Content */}
        <main className="flex-1 p-8 overflow-y-auto">
          <header className="flex justify-between items-center mb-8">
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
          </header>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {sectors.map(sector => (
              <DroppableSector
                key={sector.id}
                sector={sector}
                machines={machines.filter(m => m.group_id === sector.id)}
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
