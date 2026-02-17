// Detecta automaticamente a URL base (funciona local e na nuvem)
const API_BASE = window.location.origin + '/api';
let currentAPIKey = null;
let currentFactory = null;
let refreshInterval = null;
let efficiencyChart = null;
let productionChart = null;
let useSession = false; // true = usar cookie de sessão em vez de api_key

// ==================== NAVEGAÇÃO ====================

document.addEventListener('DOMContentLoaded', function() {
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', () => {
            navigateTo(item.dataset.page);
        });
    });
    initApp();
});

// Verifica sessão (Google) e redireciona para login, criar fábrica ou dashboard
async function initApp() {
    const hash = (window.location.hash || '').replace(/^#\/?/, '');
    try {
        const res = await fetch(`${API_BASE}/auth/me`, { credentials: 'include' });
        if (!res.ok) {
            showLoginGoogle();
            if (hash === 'dashboard') return;
            if (document.getElementById('api-key-input').value.trim()) loadDashboard();
            return;
        }
        const data = await res.json();
        const user = data.user;
        const factory = data.factory;
        if (factory) {
            useSession = true;
            currentFactory = { name: factory.name, id: factory.id };
            enterDashboard();
            return;
        }
        document.getElementById('login-google-card').style.display = 'none';
        document.getElementById('login-create-factory-card').style.display = 'block';
    } catch (e) {
        showLoginGoogle();
        if (document.getElementById('api-key-input').value.trim()) loadDashboard();
    }
}

function showLoginGoogle() {
    document.getElementById('login-google-card').style.display = 'block';
    document.getElementById('login-create-factory-card').style.display = 'none';
}

function enterDashboard() {
    document.getElementById('login-section').classList.remove('active');
    document.getElementById('live-page').classList.add('active');
    document.body.classList.remove('logged-out');
    document.getElementById('sidebar').style.display = 'flex';
    document.getElementById('sidebar-factory-name').textContent = currentFactory?.name || '-';
    if (currentFactory) {
        document.getElementById('config-factory-name').textContent = currentFactory.name;
        document.getElementById('config-api-key').textContent = '•••••••• (use Regenerar se perdeu)';
    }
    document.getElementById('config-endpoint').textContent = `${window.location.origin}/api/ingest`;
    document.getElementById('btn-regenerate-api').style.display = useSession ? 'inline-block' : 'none';
    fetchDashboardData();
    if (refreshInterval) clearInterval(refreshInterval);
    refreshInterval = setInterval(() => {
        fetchDashboardData();
        const active = document.querySelector('.nav-item.active')?.dataset.page;
        if (active === 'financeiro' || active === 'comparativo') loadAnalytics();
    }, 2000);
}

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

    if (page === 'financeiro' || page === 'comparativo') {
        loadAnalytics();
    } else if (page === 'ativos') {
        loadAtivosPage();
    } else if (page === 'ia') {
        loadIAPage();
    } else if (page === 'nxd-ativos') {
        loadNXDAtivosPage();
    }
}

// ==================== AUTENTICAÇÃO ====================

async function createFactoryAuth() {
    const name = document.getElementById('factory-name').value.trim();
    if (!name) {
        alert('Digite o nome da fábrica');
        return;
    }
    try {
        const res = await fetch(`${API_BASE}/factory`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ name })
        });
        if (!res.ok) {
            const err = await res.text();
            throw new Error(err || 'Erro ao criar fábrica');
        }
        const data = await res.json();
        const resultDiv = document.getElementById('api-key-result');
        resultDiv.innerHTML = `
            <h4>✅ Fábrica criada</h4>
            <p><strong>API Key (mostrada uma vez):</strong></p>
            <div class="api-key-display">${data.api_key}</div>
            <p><small>Copie e configure no nanoDX. Se perder, use "Regenerar API Key" nas Configurações.</small></p>
            <button class="btn-primary" onclick="location.reload()">Ir para o dashboard</button>
        `;
        resultDiv.style.display = 'block';
        currentFactory = { name: data.name, id: data.id };
        useSession = true;
    } catch (error) {
        alert('Erro: ' + error.message);
    }
}

async function createFactory() {
    const name = document.getElementById('factory-name').value.trim();
    if (!name) { alert('Digite o nome da fábrica'); return; }
    try {
        const res = await fetch(`${API_BASE}/factory/create`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name })
        });
        if (!res.ok) throw new Error('Erro ao criar fábrica');
        const data = await res.json();
        document.getElementById('api-key-result').innerHTML = `
            <h4>✅ Fábrica criada</h4>
            <div class="api-key-display">${data.api_key}</div>
            <p><small>Configure no DX: ${window.location.origin}/api/ingest</small></p>
        `;
        document.getElementById('api-key-result').style.display = 'block';
        document.getElementById('api-key-input').value = data.api_key;
    } catch (error) {
        alert('Erro: ' + error.message);
    }
}

async function loadDashboard() {
    const apiKey = document.getElementById('api-key-input').value.trim();
    if (!apiKey) {
        alert('Digite a API Key ou entre com Google');
        return;
    }
    currentAPIKey = apiKey;
    useSession = false;
    try {
        await fetchDashboardData();
        document.getElementById('login-section').classList.remove('active');
        document.getElementById('live-page').classList.add('active');
        document.body.classList.remove('logged-out');
        document.getElementById('sidebar').style.display = 'flex';
        document.getElementById('config-api-key').textContent = apiKey;
        document.getElementById('config-endpoint').textContent = `${window.location.origin}/api/ingest`;
        if (refreshInterval) clearInterval(refreshInterval);
        refreshInterval = setInterval(() => {
            fetchDashboardData();
            const activePage = document.querySelector('.nav-item.active')?.dataset.page;
            if (activePage === 'financeiro' || activePage === 'comparativo') loadAnalytics();
        }, 2000);
    } catch (error) {
        alert('Erro ao carregar: ' + error.message);
    }
}

function logout() {
    if (refreshInterval) clearInterval(refreshInterval);
    refreshInterval = null;
    currentAPIKey = null;
    currentFactory = null;
    useSession = false;
    document.body.classList.add('logged-out');
    document.getElementById('sidebar').style.display = 'none';
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    document.getElementById('login-section').classList.add('active');
    document.getElementById('api-key-input').value = '';
    showLoginGoogle();
    document.getElementById('login-create-factory-card').style.display = 'none';
    window.location.href = window.location.pathname;
}

async function regenerateApiKey() {
    if (!useSession) return;
    if (!confirm('Gerar nova API Key? A antiga deixará de funcionar. Configure a nova no nanoDX.')) return;
    try {
        const res = await fetch(`${API_BASE}/factory/regenerate-api-key`, { method: 'POST', credentials: 'include' });
        if (!res.ok) throw new Error(await res.text());
        const data = await res.json();
        document.getElementById('config-api-key').textContent = data.api_key;
        alert('Nova API Key gerada. Copie e configure no nanoDX. Esta é a única vez que ela será exibida.');
    } catch (e) {
        alert('Erro: ' + e.message);
    }
}

// ==================== LIVE VIEW ====================

async function fetchDashboardData() {
    try {
        const url = useSession ? `${API_BASE}/dashboard` : `${API_BASE}/dashboard?api_key=${encodeURIComponent(currentAPIKey)}`;
        const opts = useSession ? { credentials: 'include' } : {};
        const response = await fetch(url, opts);
        if (!response.ok) throw new Error(useSession ? 'Sessão inválida' : 'API Key inválida');

        const data = await response.json();
        currentFactory = data.factory;
        
        document.getElementById('sidebar-factory-name').textContent = data.factory.name;
        document.getElementById('config-factory-name').textContent = data.factory.name;
        document.getElementById('config-machine-count').textContent = data.machines?.length || 0;
        
        // Dashboard de Alívio: resumo (cards)
        fetchAlivioSummary();
        
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

async function loadAtivosPage() {
    const sectorsUrl = useSession ? `${API_BASE}/sectors` : `${API_BASE}/sectors?api_key=${encodeURIComponent(currentAPIKey)}`;
    const opts = useSession ? { credentials: 'include' } : {};
    try {
        const [secRes, dashRes] = await Promise.all([
            fetch(sectorsUrl, opts),
            fetch(useSession ? `${API_BASE}/dashboard` : `${API_BASE}/dashboard?api_key=${encodeURIComponent(currentAPIKey)}`, opts)
        ]);
        const sectors = secRes.ok ? (await secRes.json()).sectors || [] : [];
        const dash = dashRes.ok ? await dashRes.json() : { machines: [] };
        const machines = dash.machines || [];
        const listEl = document.getElementById('sectors-list');
        listEl.innerHTML = sectors.length ? sectors.map(s => `<li><strong>${s.name}</strong></li>`).join('') : '<li>Nenhum setor ainda. Crie um acima.</li>';
        const wrap = document.getElementById('ativos-machines');
        if (!machines.length) {
            wrap.innerHTML = '<p class="empty-state">Nenhuma máquina. Configure o DX e envie dados.</p>';
            return;
        }
        wrap.innerHTML = `
            <table class="data-table">
                <thead><tr><th>Máquina</th><th>Nome (exibição)</th><th>Setor</th><th>Anotações</th><th>Ação</th></tr></thead>
                <tbody>
                    ${machines.map(m => `
                        <tr data-machine-id="${m.id}">
                            <td>${m.name}</td>
                            <td><input type="text" class="asset-display-name" value="${(m.display_name || '').replace(/"/g, '&quot;')}" placeholder="Nome que você entende" /></td>
                            <td>
                                <select class="asset-sector">
                                    <option value="0">-- Nenhum --</option>
                                    ${sectors.map(s => `<option value="${s.id}" ${(m.sector_id === s.id || (m.sector_id && m.sector_id === s.id)) ? 'selected' : ''}>${s.name}</option>`).join('')}
                                </select>
                            </td>
                            <td><input type="text" class="asset-notes" value="${(m.notes || '').replace(/"/g, '&quot;')}" placeholder="Anotações" /></td>
                            <td><button class="btn-secondary btn-sm" onclick="saveMachineAsset(${m.id})">Salvar</button></td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
    } catch (e) {
        document.getElementById('ativos-machines').innerHTML = '<p class="empty-state">Erro ao carregar.</p>';
    }
}

async function createSector() {
    const name = document.getElementById('sector-name').value.trim();
    if (!name) { alert('Digite o nome do setor'); return; }
    const url = useSession ? `${API_BASE}/sectors` : `${API_BASE}/sectors?api_key=${encodeURIComponent(currentAPIKey)}`;
    const opts = { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name }), credentials: useSession ? 'include' : 'omit' };
    try {
        const res = await fetch(url, opts);
        if (!res.ok) throw new Error(await res.text());
        document.getElementById('sector-name').value = '';
        loadAtivosPage();
    } catch (e) {
        alert('Erro: ' + e.message);
    }
}

async function saveMachineAsset(machineId) {
    const row = document.querySelector(`tr[data-machine-id="${machineId}"]`);
    if (!row) return;
    const displayName = row.querySelector('.asset-display-name').value.trim();
    const notes = row.querySelector('.asset-notes').value.trim();
    const sectorId = parseInt(row.querySelector('.asset-sector').value, 10) || 0;
    const url = useSession ? `${API_BASE}/machine/asset` : `${API_BASE}/machine/asset?api_key=${encodeURIComponent(currentAPIKey)}`;
    const opts = { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ machine_id: machineId, display_name: displayName, notes, sector_id: sectorId }), credentials: useSession ? 'include' : 'omit' };
    try {
        const res = await fetch(url, opts);
        if (!res.ok) throw new Error(await res.text());
        if (currentFactory) document.getElementById('config-machine-count').textContent = document.querySelectorAll('#ativos-machines tbody tr').length;
    } catch (e) {
        alert('Erro: ' + e.message);
    }
}

async function loadIAPage() {
    const url = useSession ? `${API_BASE}/sectors` : `${API_BASE}/sectors?api_key=${encodeURIComponent(currentAPIKey)}`;
    try {
        const res = await fetch(url, useSession ? { credentials: 'include' } : {});
        const data = res.ok ? (await res.json()).sectors || [] : [];
        const sel = document.getElementById('ia-sector');
        sel.innerHTML = '<option value="">Todos</option>' + data.map(s => `<option value="${s.id}">${s.name}</option>`).join('');
    } catch (_) {}
}

async function generateReportIA() {
    const sectorId = document.getElementById('ia-sector').value;
    const period = document.getElementById('ia-period').value;
    const shift = document.getElementById('ia-shift').value;
    const detail = document.getElementById('ia-detail').value;
    const niches = [];
    if (document.getElementById('ia-nicho-financeiro').checked) niches.push('financeiro');
    if (document.getElementById('ia-nicho-producao').checked) niches.push('producao');
    if (document.getElementById('ia-nicho-manutencao').checked) niches.push('manutencao');
    const url = useSession ? `${API_BASE}/report/ia` : `${API_BASE}/report/ia?api_key=${encodeURIComponent(currentAPIKey)}`;
    const body = { sector_id: sectorId ? parseInt(sectorId, 10) : 0, period, shift, detail_level: detail, niches: niches.length ? niches : ['financeiro', 'producao', 'manutencao'] };
    try {
        document.getElementById('ia-report-text').textContent = 'Gerando...';
        document.getElementById('ia-result').style.display = 'block';
        const res = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
            credentials: useSession ? 'include' : 'omit'
        });
        if (!res.ok) throw new Error(await res.text());
        const data = await res.json();
        document.getElementById('ia-report-text').innerHTML = (data.report || '').replace(/\n/g, '<br>');
        const rast = (data.rastreabilidade || []).map(r =>
            `<div class="rastreabilidade-item"><strong>${r.metrica || ''}</strong><br>Fonte: ${r.fonte || ''} | Origem: ${r.origem || ''} | Processamento: ${r.processamento || ''} | ${r.data_hora || ''}</div>`
        ).join('');
        document.getElementById('ia-rastreabilidade').innerHTML = rast || '<p>Nenhum item de rastreabilidade.</p>';
    } catch (e) {
        document.getElementById('ia-report-text').textContent = 'Erro: ' + e.message;
    }
}

// ==================== NXD v2 Gestão de Ativos ====================
const NXD_BASE = window.location.origin + '/nxd';
let nxdSelectedFactoryId = null;

async function loadNXDAtivosPage() {
    const content = document.getElementById('nxd-ativos-content');
    const empty = document.getElementById('nxd-ativos-empty');
    const sel = document.getElementById('nxd-factory-select');
    const createBtn = document.getElementById('nxd-create-factory-btn');
    try {
        const res = await fetch(`${NXD_BASE}/factories`, { credentials: 'include' });
        if (!res.ok) {
            empty.style.display = 'block';
            content.style.display = 'none';
            empty.textContent = 'NXD não disponível ou faça login.';
            return;
        }
        const data = await res.json();
        const factories = data.factories || [];
        sel.innerHTML = '<option value="">-- Selecione --</option>' + factories.map(f => `<option value="${f.id}">${f.name}</option>`).join('');
        if (factories.length === 0) {
            createBtn.style.display = 'inline-block';
            createBtn.onclick = createNXDFactory;
            empty.style.display = 'block';
            content.style.display = 'none';
            empty.textContent = 'Nenhuma fábrica NXD. Clique em "Criar fábrica NXD".';
            return;
        }
        createBtn.style.display = 'none';
        empty.style.display = 'none';
        content.style.display = 'grid';
        sel.onchange = onNXDFactoryChange;
        if (sel.value) onNXDFactoryChange();
    } catch (e) {
        empty.style.display = 'block';
        content.style.display = 'none';
        empty.textContent = 'Erro ao carregar NXD.';
    }
}

async function createNXDFactory() {
    const name = prompt('Nome da fábrica NXD:');
    if (!name) return;
    try {
        const res = await fetch(`${NXD_BASE}/factories`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ name })
        });
        if (!res.ok) throw new Error(await res.text());
        loadNXDAtivosPage();
    } catch (e) {
        alert('Erro: ' + e.message);
    }
}

function onNXDFactoryChange() {
    const sel = document.getElementById('nxd-factory-select');
    nxdSelectedFactoryId = sel.value || null;
    if (!nxdSelectedFactoryId) return;
    loadNXDGroupsAndAssets();
}

async function loadNXDGroupsAndAssets() {
    if (!nxdSelectedFactoryId) return;
    const q = `?factory_id=${encodeURIComponent(nxdSelectedFactoryId)}`;
    try {
        const [gRes, aRes] = await Promise.all([
            fetch(`${NXD_BASE}/groups${q}`, { credentials: 'include' }),
            fetch(`${NXD_BASE}/assets${q}&ungrouped=true`, { credentials: 'include' })
        ]);
        const groups = gRes.ok ? (await gRes.json()).groups || [] : [];
        const ungrouped = aRes.ok ? (await aRes.json()).assets || [] : [];
        const search = document.getElementById('nxd-search-ungrouped').value.trim();
        let filtered = ungrouped;
        if (search) {
            const s = search.toLowerCase();
            filtered = ungrouped.filter(a => (a.display_name || '').toLowerCase().includes(s) || (a.source_tag_id || '').toLowerCase().includes(s));
        }
        document.getElementById('nxd-ungrouped-list').innerHTML = filtered.map(a =>
            `<li class="nxd-asset-item" data-id="${a.id}">
                <span class="nxd-asset-name">${a.display_name || a.source_tag_id}</span>
                <select class="nxd-move-select" onchange="nxdMoveAsset('${a.id}', this.value)">
                    <option value="">-- Mover para --</option>
                    ${groups.map(g => `<option value="${g.id}">${g.name}</option>`).join('')}
                </select>
            </li>`
        ).join('');
        const cardsEl = document.getElementById('nxd-groups-cards');
        cardsEl.innerHTML = groups.map(g => `
            <div class="nxd-group-card" data-group-id="${g.id}">
                <div class="nxd-group-header">
                    <input type="text" class="nxd-group-name" value="${(g.name || '').replace(/"/g, '&quot;')}" onchange="nxdRenameGroup('${g.id}', this.value)" />
                </div>
                <ul class="nxd-group-assets" id="nxd-group-${g.id}"></ul>
            </div>
        `).join('');
        for (const g of groups) {
            const ar = await fetch(`${NXD_BASE}/assets${q}&group_id=${g.id}`, { credentials: 'include' });
            const arr = ar.ok ? (await ar.json()).assets || [] : [];
            const ul = document.getElementById(`nxd-group-${g.id}`);
            if (ul) ul.innerHTML = arr.map(a =>
                `<li class="nxd-asset-item"><span>${a.display_name || a.source_tag_id}</span>
                 <button class="btn-sm" onclick="nxdMoveAsset('${a.id}', '')">Soltar</button></li>`
            ).join('');
        }
    } catch (e) {
        console.error(e);
    }
}

document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.getElementById('nxd-search-ungrouped');
    if (searchInput) searchInput.addEventListener('input', function() {
        if (nxdSelectedFactoryId) loadNXDGroupsAndAssets();
    });
    const createGroupBtn = document.getElementById('nxd-create-group-btn');
    if (createGroupBtn) createGroupBtn.addEventListener('click', async function() {
        const name = prompt('Nome do setor:');
        if (!name || !nxdSelectedFactoryId) return;
        try {
            const res = await fetch(`${NXD_BASE}/groups?factory_id=${encodeURIComponent(nxdSelectedFactoryId)}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ name })
            });
            if (!res.ok) throw new Error(await res.text());
            loadNXDGroupsAndAssets();
        } catch (e) {
            alert('Erro: ' + e.message);
        }
    });
});

async function nxdMoveAsset(assetId, groupId) {
    if (!nxdSelectedFactoryId) return;
    try {
        const res = await fetch(`${NXD_BASE}/assets/${assetId}/move?factory_id=${encodeURIComponent(nxdSelectedFactoryId)}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ group_id: groupId || null })
        });
        if (!res.ok) throw new Error(await res.text());
        loadNXDGroupsAndAssets();
    } catch (e) {
        alert('Erro: ' + e.message);
    }
}

async function nxdRenameGroup(groupId, name) {
    if (!nxdSelectedFactoryId || !name) return;
    try {
        const res = await fetch(`${NXD_BASE}/groups/${groupId}?factory_id=${encodeURIComponent(nxdSelectedFactoryId)}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ name })
        });
        if (!res.ok) throw new Error(await res.text());
    } catch (e) {
        alert('Erro: ' + e.message);
    }
}

async function fetchAlivioSummary() {
    try {
        const url = useSession ? `${API_BASE}/dashboard/summary` : `${API_BASE}/dashboard/summary?api_key=${encodeURIComponent(currentAPIKey)}`;
        const opts = useSession ? { credentials: 'include' } : {};
        const res = await fetch(url, opts);
        if (!res.ok) return;
        const d = await res.json();
        document.getElementById('card-online').textContent = d.online ?? 0;
        document.getElementById('card-offline').textContent = d.offline ?? 0;
        document.getElementById('card-critical').textContent = d.critical ?? 0;
        document.getElementById('card-pecas').textContent = (d.total_pecas ?? 0).toLocaleString('pt-BR');
        document.getElementById('card-lucro').textContent = 'R$ ' + (d.lucro_cessante ?? 0).toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
        const msgEl = document.getElementById('alivio-message');
        if (d.message) { msgEl.textContent = d.message; msgEl.style.display = 'block'; } else { msgEl.style.display = 'none'; }
    } catch (e) { /* ignore */ }
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
        const analyticsUrl = useSession ? `${API_BASE}/analytics` : `${API_BASE}/analytics?api_key=${encodeURIComponent(currentAPIKey)}`;
        const response = await fetch(analyticsUrl, useSession ? { credentials: 'include' } : {});
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

async function initApp() {
    // Começa na tela de login
    document.body.classList.add('logged-out');
    const sidebar = document.getElementById('sidebar');
    if (sidebar) {
        sidebar.style.display = 'none';
    }
    
    try {
        const response = await fetch(`${API_BASE}/health`);
        if (response.ok) {
            console.log('✓ Servidor NXD online');
        }
    } catch (error) {
        console.error('❌ Servidor offline:', error);
    }
}
