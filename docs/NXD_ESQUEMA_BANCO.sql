-- =============================================================================
-- NXD (Nexus Data Exchange) — Esquema do Banco de Dados
-- Estrutura: Unidades (Fábricas) → Setores (Baias) → Máquinas → Tags → Logs
-- Rastreabilidade: cada dado tem fonte (data_points.id), origem (PLC/Tag), processamento e timestamp
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. UNIDADES (Fábricas / Contas)
-- Uma unidade = uma fábrica ou planta. 1 assinatura = 1 fábrica (modelo atual).
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS factories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,                    -- dono (Google); NULL = legado/API key apenas
    name TEXT NOT NULL,
    api_key TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT 1,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_factories_user ON factories(user_id);

-- -----------------------------------------------------------------------------
-- 2. SETORES (Baias)
-- O gestor cria setores (ex: "Linha de Montagem A") e agrupa máquinas.
-- Pesquisável por nome; usado na Central de IA para filtrar relatórios.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sectors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    factory_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (factory_id) REFERENCES factories(id)
);
CREATE INDEX IF NOT EXISTS idx_sectors_factory ON sectors(factory_id);

-- -----------------------------------------------------------------------------
-- 3. MÁQUINAS
-- Auto-discovery: criadas quando o DX envia pela primeira vez (device_id + api_key).
-- display_name e notes: o gestor renomeia e anota na tela "Ativos / Baias".
-- Setor: atribuído via machine_sector (N:1 máquina → setor).
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS machines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    factory_id INTEGER NOT NULL,
    device_id TEXT NOT NULL,
    name TEXT,                          -- nome técnico (ex: DX-001)
    display_name TEXT,                  -- nome que o gestor entende (ex: "Inj 01")
    notes TEXT,                         -- anotações do gestor
    brand TEXT,
    protocol TEXT,
    last_seen DATETIME,
    status TEXT DEFAULT 'offline',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (factory_id) REFERENCES factories(id),
    UNIQUE(factory_id, device_id)
);
CREATE INDEX IF NOT EXISTS idx_machines_factory ON machines(factory_id);

-- Relação máquina → setor (1 máquina = 1 setor; opcional)
CREATE TABLE IF NOT EXISTS machine_sector (
    machine_id INTEGER PRIMARY KEY,
    sector_id INTEGER NOT NULL,
    FOREIGN KEY (machine_id) REFERENCES machines(id),
    FOREIGN KEY (sector_id) REFERENCES sectors(id)
);

-- -----------------------------------------------------------------------------
-- 4. TAGS
-- Auto-discovery: criadas quando o DX envia uma tag nova (tag_name por machine_id).
-- Representam os pontos lidos do CLP/PLC (ex: Total_Pecas, Status_Producao).
-- Rastreabilidade: "Origem" = máquina + tag_name.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    machine_id INTEGER NOT NULL,
    tag_name TEXT NOT NULL,
    tag_type TEXT NOT NULL,             -- float, int, bool, string
    unit TEXT,
    min_value REAL,
    max_value REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (machine_id) REFERENCES machines(id),
    UNIQUE(machine_id, tag_name)
);
CREATE INDEX IF NOT EXISTS idx_tags_machine ON tags(machine_id);

-- -----------------------------------------------------------------------------
-- 5. LOGS DE DADOS (data_points) — FONTE DA VERDADE
-- Cada linha = uma leitura do DX: qual máquina, qual tag, valor, timestamp.
-- Rastreabilidade:
--   Fonte    = id deste registro (log bruto do hardware DX)
--   Origem   = machines.name + tags.tag_name (qual PLC e qual Tag)
--   Processamento = definido no relatório (ex: "Leitura direta", "Cálculo OEE")
--   Data/Hora = timestamp
-- Histórico de longo prazo: retenção configurável (ex: 15 anos na nuvem).
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS data_points (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    machine_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    value TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    quality TEXT DEFAULT 'good',
    FOREIGN KEY (machine_id) REFERENCES machines(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);
CREATE INDEX IF NOT EXISTS idx_data_points_timestamp ON data_points(timestamp);
CREATE INDEX IF NOT EXISTS idx_data_points_machine ON data_points(machine_id);
CREATE INDEX IF NOT EXISTS idx_data_points_tag ON data_points(tag_id);

-- -----------------------------------------------------------------------------
-- 6. ALERTAS (regras por tag)
-- Condições configuráveis por tag (ex: temperatura > limite).
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tag_id INTEGER NOT NULL,
    condition TEXT NOT NULL,
    threshold REAL NOT NULL,
    message TEXT,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);

-- -----------------------------------------------------------------------------
-- 7. LOGS DE AUDITORIA
-- Ações de ingestão, login, erros. Suporte e diagnóstico.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    action TEXT NOT NULL,
    api_key TEXT,
    device_id TEXT,
    status TEXT NOT NULL,
    message TEXT,
    ip_address TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp);

-- -----------------------------------------------------------------------------
-- 8. USUÁRIOS (Google Auth)
-- 1 usuário pode ter 1 fábrica (modelo atual). user_id em factories.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    google_uid TEXT UNIQUE NOT NULL,
    terms_accepted_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_google_uid ON users(google_uid);

-- =============================================================================
-- RESUMO DO CAMINHO DA VERDADE (Rastreabilidade)
-- =============================================================================
-- Para qualquer métrica no relatório da Central de IA:
--   1. Fonte      → data_points.id (ou "agregado NXD" quando for soma/média)
--   2. Origem     → machines.name + tags.tag_name (PLC e Tag lida)
--   3. Processamento → ex: "Leitura direta", "Cálculo: (5/60)*Custo_Hora_Parada"
--   4. Data/Hora  → data_points.timestamp (ou timestamp da geração do agregado)
-- Assim o dado não é "opinião da IA", é fato rastreável e auditável.
