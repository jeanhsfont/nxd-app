// Detecta automaticamente a URL base (funciona local e na nuvem)
const API_BASE = window.location.origin + '/api';
let currentAPIKey = null;
let currentFactory = null;
let refreshInterval = null;
let efficiencyChart = null;
let productionChart = null;

// ==================== NAVEGAÇÃO ====================

document.querySelectorAll('.nav-item').forEach(item => {
    item.addEventListener('click', () => {
        const page = item.dataset.page;
        navigateTo(page);
    });
});

function navigateTo(page) {
    // Atualiza menu
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
        if (item.dataset.page === page) {
            item.classList.add('active');
        }
    });

    // Atualiza páginas
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    document.getElementById(`${page}-page`).classList.add('active');

    // Carrega dados específicos da página
    if (page === 'financeiro') {
        loadAnalytics();
    } else if (page === 'comparativo') {
        loadAnalytics();
    }
}

// ==================== AUTENTICAÇÃO ====================

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
            <p><small>⚠️ Guarde esta chave! Configure no DX: ${window.location.origin}/api/ingest</small></p>
        `;
        resultDiv.style.display = 'block';
        
        document.getElementById('factory-name').value = '';
        document.getElementById('api-key-input').value = data.api_key;
    } catch (error) {
        alert('Erro ao criar fábrica: ' + error.message);
    }
}

async function loadDashboard() {
    const apiKey = document.getElementById('api-key-input').value.trim();
    if (!apiKey) {
        alert('Por favor, digite a API Key');
        return;
    }

    currentAPIKey = apiKey;
    
    try {
        await fetchDashboardData();
        
        // Esconde login, mostra sidebar
        document.getElementById('login-section').classList.remove('active');
        document.getElementById('live-page').classList.add('active');
        document.body.classList.remove('logged-out');
        document.getElementById('sidebar').style.display = 'flex';

        // Configura página de config
        document.getElementById('config-api-key').textContent = apiKey;
        document.getElementById('config-endpoint').textContent = `${window.location.origin}/api/ingest`;
        
        // Atualiza a cada 2 segundos
        if (refreshInterval) clearInterval(refreshInterval);
        refreshInterval = setInterval(() => {
            fetchDashboardData();
            const activePage = document.querySelector('.nav-item.active')?.dataset.page;
            if (activePage === 'financeiro' || activePage === 'comparativo') {
                loadAnalytics();
            }
        }, 2000);
    } catch (error) {
        alert('Erro ao carregar dashboard: ' + error.message);
    }
}

function logout() {
    if (refreshInterval) clearInterval(refreshInterval);
    currentAPIKey = null;
    currentFactory = null;
    
    document.body.classList.add('logged-out');
    document.getElementById('sidebar').style.display = 'none';
    
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    document.getElementById('login-section').classList.add('active');
    document.getElementById('api-key-input').value = '';
}

// ==================== LIVE VIEW ====================

async function fetchDashboardData() {
    try {
        const response = await fetch(`${API_BASE}/dashboard?api_key=${currentAPIKey}`);
        if (!response.ok) throw new Error('API Key inválida');

        const data = await response.json();
        currentFactory = data.factory;
        
        // Atualiza nome da fábrica
        document.getElementById('sidebar-factory-name').textContent = data.factory.name;
        document.getElementById('config-factory-name').textContent = data.factory.name;
        document.getElementById('config-machine-count').textContent = data.machines?.length || 0;
        
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

function renderMachines(machines) {
    const grid = document.getElementById('machines-grid');
    grid.innerHTML = '';

    machines.forEach(machine => {
        const card = document.createElement('div');
        card.className = 'machine-card';
        
        // Detecta status baseado na tag Status_Producao
        const statusValue = machine.values?.['Status_Producao'];
        const isOnline = statusValue === 'true' || statusValue === '1';
        const statusClass = isOnline ? 'status-online' : 'status-offline';
        const statusText = isOnline ? '🟢 Produzindo' : '🔴 Parado';
        
        let tagsHTML = '';
        if (machine.tags && machine.tags.length > 0) {
            machine.tags.forEach(tag => {
                let value = machine.values?.[tag.tag_name] || 'N/A';
                
                // Formata valores
                if (value === 'true') value = '✓ Sim';
                else if (value === 'false') value = '✗ Não';
                else if (!isNaN(parseFloat(value)) && value.includes('.')) {
                    value = parseFloat(value).toFixed(1);
                }
                
                tagsHTML += `
                    <div class="tag-item">
                        <span class="tag-name">${formatTagName(tag.tag_name)}</span>
                        <span class="tag-value">${value}</span>
                    </div>
                `;
            });
        } else {
            tagsHTML = '<p style="text-align: center; color: #6b7280;">Aguardando dados...</p>';
        }

        const lastSeen = machine.last_seen ? new Date(machine.last_seen).toLocaleString('pt-BR') : '-';

        card.innerHTML = `
            <div class="machine-header">
                <div class="machine-name">${formatMachineName(machine.name)}</div>
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

// ==================== ANALYTICS ====================

async function loadAnalytics() {
    try {
        const response = await fetch(`${API_BASE}/analytics?api_key=${currentAPIKey}`);
        if (!response.ok) return;

        const data = await response.json();
        
        // Atualiza página Financeiro
        updateFinanceiroPage(data);
        
        // Atualiza página Comparativo
        updateComparativoPage(data);
        
    } catch (error) {
        console.error('Erro ao carregar analytics:', error);
    }
}

function updateFinanceiroPage(data) {
    // Lucro Cessante Total
    const lucroTotal = data.machines?.reduce((sum, m) => sum + (m.lucro_cessante || 0), 0) || 0;
    document.getElementById('lucro-cessante-total').textContent = formatCurrency(lucroTotal);
    
    // Detalhes por máquina
    const lucroDetails = document.getElementById('lucro-details');
    lucroDetails.innerHTML = '';
    
    data.machines?.forEach(machine => {
        const statusClass = machine.status ? 'online' : 'offline';
        const statusText = machine.status ? 'Produzindo' : 'Parado';
        
        lucroDetails.innerHTML += `
            <div class="lucro-item">
                <div class="lucro-machine">
                    <div class="lucro-status ${statusClass}"></div>
                    <div>
                        <strong>${formatMachineName(machine.name)}</strong>
                        <br><small>${statusText}</small>
                    </div>
                </div>
                <div class="lucro-value">${machine.lucro_cessante > 0 ? '-' : ''} ${formatCurrency(machine.lucro_cessante)}</div>
            </div>
        `;
    });
    
    // Total de peças
    document.getElementById('total-pecas').textContent = formatNumber(data.total_pecas || 0);
    
    // Total energia
    const totalEnergia = data.machines?.reduce((sum, m) => sum + (m.consumo_energia || 0), 0) || 0;
    document.getElementById('total-energia').textContent = totalEnergia.toFixed(1);
    
    // Tabela financeira
    const tbody = document.getElementById('financial-table-body');
    tbody.innerHTML = '';
    
    data.machines?.forEach(machine => {
        const statusClass = machine.status ? 'status-online' : 'status-offline';
        const statusText = machine.status ? '🟢 Online' : '🔴 Parado';
        
        tbody.innerHTML += `
            <tr>
                <td><strong>${formatMachineName(machine.name)}</strong> (${machine.brand})</td>
                <td><span class="status-badge ${statusClass}">${statusText}</span></td>
                <td>${formatCurrency(machine.custo_hora_parada)}/hora</td>
                <td>${machine.tempo_parado_min?.toFixed(0) || 0} min</td>
                <td style="color: ${machine.lucro_cessante > 0 ? 'var(--danger)' : 'var(--gray-500)'}; font-weight: 600;">
                    ${machine.lucro_cessante > 0 ? '-' : ''} ${formatCurrency(machine.lucro_cessante)}
                </td>
            </tr>
        `;
    });
}

function updateComparativoPage(data) {
    // Gráfico de Eficiência
    updateEfficiencyChart(data.machines || []);
    
    // Indicadores de Saúde
    updateHealthIndicators(data.machines || []);
    
    // Gráfico de Produção
    updateProductionChart(data.machines || []);
    
    // Vencedor de eficiência
    const winner = document.getElementById('efficiency-winner');
    if (data.mais_eficiente && data.ganho_eficiencia > 0) {
        winner.innerHTML = `
            <div class="trophy">🏆</div>
            <h4>${formatMachineName(data.mais_eficiente)}</h4>
            <p>Mais eficiente: <strong>${data.ganho_eficiencia.toFixed(1)}%</strong> melhor que a segunda</p>
        `;
        winner.style.display = 'block';
    } else {
        winner.style.display = 'none';
    }
}

function updateEfficiencyChart(machines) {
    const ctx = document.getElementById('efficiency-chart')?.getContext('2d');
    if (!ctx) return;

    const labels = machines.map(m => formatMachineName(m.name));
    const values = machines.map(m => m.eficiencia || 0);
    const colors = machines.map((m, i) => i === 0 ? '#6366f1' : '#10b981');

    if (efficiencyChart) {
        efficiencyChart.data.labels = labels;
        efficiencyChart.data.datasets[0].data = values;
        efficiencyChart.data.datasets[0].backgroundColor = colors;
        efficiencyChart.update();
    } else {
        efficiencyChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Eficiência (peças/kWh)',
                    data: values,
                    backgroundColor: colors,
                    borderRadius: 8
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Peças por kWh'
                        }
                    }
                }
            }
        });
    }
}

function updateHealthIndicators(machines) {
    const container = document.getElementById('health-indicators');
    container.innerHTML = '';

    machines.forEach(machine => {
        const health = machine.health_score || 0;
        let colorClass = 'green';
        let statusClass = 'good';
        let statusText = 'Saudável';

        if (health < 70) {
            colorClass = 'red';
            statusClass = 'critical';
            statusText = 'Crítico';
        } else if (health < 85) {
            colorClass = 'yellow';
            statusClass = 'warning';
            statusText = 'Atenção';
        }

        container.innerHTML += `
            <div class="health-item">
                <div class="health-gauge ${colorClass}" style="--value: ${health}">
                    <span>${health.toFixed(0)}%</span>
                </div>
                <div class="health-info">
                    <h4>${formatMachineName(machine.name)}</h4>
                    <p>Temperatura: ${machine.temperatura?.toFixed(1) || '-'}°C</p>
                </div>
                <span class="health-status ${statusClass}">${statusText}</span>
            </div>
        `;
    });
}

function updateProductionChart(machines) {
    const ctx = document.getElementById('production-chart')?.getContext('2d');
    if (!ctx) return;

    const labels = machines.map(m => formatMachineName(m.name));
    const pecas = machines.map(m => m.total_pecas || 0);
    const energia = machines.map(m => (m.consumo_energia || 0) * 10); // Escala para visualização

    if (productionChart) {
        productionChart.data.labels = labels;
        productionChart.data.datasets[0].data = pecas;
        productionChart.data.datasets[1].data = energia;
        productionChart.update();
    } else {
        productionChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [
                    {
                        label: 'Peças Produzidas',
                        data: pecas,
                        backgroundColor: '#6366f1',
                        borderRadius: 8
                    },
                    {
                        label: 'Energia (kWh x10)',
                        data: energia,
                        backgroundColor: '#f59e0b',
                        borderRadius: 8
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { position: 'top' }
                },
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }
}

// ==================== UTILS ====================

function formatTagName(name) {
    return name.replace(/_/g, ' ');
}

function formatMachineName(name) {
    if (!name) return '-';
    // Remove prefixos comuns
    return name
        .replace('INJETORA_', '')
        .replace('_01', ' #1')
        .replace('_02', ' #2')
        .replace(/_/g, ' ');
}

function formatCurrency(value) {
    return new Intl.NumberFormat('pt-BR', {
        style: 'currency',
        currency: 'BRL'
    }).format(value || 0);
}

function formatNumber(value) {
    return new Intl.NumberFormat('pt-BR').format(value || 0);
}

function copyApiKey() {
    const apiKey = document.getElementById('config-api-key').textContent;
    navigator.clipboard.writeText(apiKey);
    alert('API Key copiada!');
}

function copyEndpoint() {
    const endpoint = document.getElementById('config-endpoint').textContent;
    navigator.clipboard.writeText(endpoint);
    alert('Endpoint copiado!');
}

// ==================== INIT ====================

window.addEventListener('load', async () => {
    // Começa na tela de login
    document.body.classList.add('logged-out');
    document.getElementById('sidebar').style.display = 'none';
    
    try {
        const response = await fetch(`${API_BASE}/health`);
        if (response.ok) {
            console.log('✓ Servidor NXD online');
        }
    } catch (error) {
        console.error('❌ Servidor offline:', error);
    }
});
