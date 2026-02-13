const API_BASE = 'http://localhost:8080/api';
let currentAPIKey = null;
let refreshInterval = null;

// Cria nova fábrica
async function createFactory() {
    const name = document.getElementById('factory-name').value.trim();
    if (!name) {
        alert('Por favor, digite o nome da fábrica');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/factory/create`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name })
        });

        if (!response.ok) throw new Error('Erro ao criar fábrica');

        const data = await response.json();
        
        const resultDiv = document.getElementById('api-key-result');
        resultDiv.innerHTML = `
            <h4>✅ Fábrica criada com sucesso!</h4>
            <p><strong>Nome:</strong> ${data.name}</p>
            <p><strong>API Key:</strong></p>
            <div class="api-key-display">${data.api_key}</div>
            <p><small>⚠️ Guarde esta chave em local seguro! Você precisará dela para configurar o módulo DX.</small></p>
        `;
        resultDiv.style.display = 'block';
        
        document.getElementById('factory-name').value = '';
        document.getElementById('api-key-input').value = data.api_key;
    } catch (error) {
        alert('Erro ao criar fábrica: ' + error.message);
    }
}

// Carrega dashboard
async function loadDashboard() {
    const apiKey = document.getElementById('api-key-input').value.trim();
    if (!apiKey) {
        alert('Por favor, digite a API Key');
        return;
    }

    currentAPIKey = apiKey;
    
    try {
        await fetchDashboardData();
        
        document.getElementById('setup-section').style.display = 'none';
        document.getElementById('dashboard-section').style.display = 'block';
        
        // Atualiza a cada 3 segundos
        if (refreshInterval) clearInterval(refreshInterval);
        refreshInterval = setInterval(fetchDashboardData, 3000);
    } catch (error) {
        alert('Erro ao carregar dashboard: ' + error.message);
    }
}

// Busca dados do dashboard
async function fetchDashboardData() {
    try {
        const response = await fetch(`${API_BASE}/dashboard?api_key=${currentAPIKey}`);
        if (!response.ok) throw new Error('API Key inválida ou fábrica não encontrada');

        const data = await response.json();
        
        document.getElementById('factory-name-display').textContent = `🏭 ${data.factory.name}`;
        
        if (!data.machines || data.machines.length === 0) {
            document.getElementById('machines-grid').style.display = 'none';
            document.getElementById('no-machines').style.display = 'block';
        } else {
            document.getElementById('machines-grid').style.display = 'grid';
            document.getElementById('no-machines').style.display = 'none';
            renderMachines(data.machines);
        }
    } catch (error) {
        console.error('Erro ao buscar dados:', error);
    }
}

// Renderiza cards de máquinas
function renderMachines(machines) {
    const grid = document.getElementById('machines-grid');
    grid.innerHTML = '';

    machines.forEach(machine => {
        const card = document.createElement('div');
        card.className = 'machine-card';
        
        const statusClass = machine.status === 'online' ? 'status-online' : 'status-offline';
        const statusText = machine.status === 'online' ? '🟢 Online' : '🔴 Offline';
        
        let tagsHTML = '';
        if (machine.tags && machine.tags.length > 0) {
            machine.tags.forEach(tag => {
                const value = machine.values[tag.tag_name] || 'N/A';
                tagsHTML += `
                    <div class="tag-item">
                        <span class="tag-name">${tag.tag_name}</span>
                        <span class="tag-value">${value}</span>
                    </div>
                `;
            });
        } else {
            tagsHTML = '<p style="text-align: center; color: #6c757d;">Aguardando dados...</p>';
        }

        const lastSeen = new Date(machine.last_seen).toLocaleString('pt-BR');

        card.innerHTML = `
            <div class="machine-header">
                <div class="machine-name">${machine.name}</div>
                <div class="machine-brand">${machine.brand}</div>
            </div>
            <div class="status-badge ${statusClass}">${statusText}</div>
            <div class="tags-list">
                ${tagsHTML}
            </div>
            <div class="last-update">
                Última atualização: ${lastSeen}
            </div>
        `;
        
        grid.appendChild(card);
    });
}

// Logout
function logout() {
    if (refreshInterval) clearInterval(refreshInterval);
    currentAPIKey = null;
    document.getElementById('setup-section').style.display = 'block';
    document.getElementById('dashboard-section').style.display = 'none';
    document.getElementById('api-key-input').value = '';
}

// Verifica health do servidor ao carregar
window.addEventListener('load', async () => {
    try {
        const response = await fetch(`${API_BASE}/health`);
        if (response.ok) {
            console.log('✓ Servidor NXD online');
        }
    } catch (error) {
        console.error('❌ Servidor offline:', error);
    }
});
