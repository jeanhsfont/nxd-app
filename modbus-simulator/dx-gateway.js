/**
 * SIMULADOR DELTA DX-2100 GATEWAY
 * 
 * Este é o coração do sistema - simula EXATAMENTE o que o DX real faz:
 * 1. Conecta nos CLPs via Modbus TCP
 * 2. Lê os registros periodicamente
 * 3. Converte para JSON
 * 4. Envia via HTTP POST para o NXD Cloud
 * 
 * A ÚNICA diferença para o DX real é:
 * - DX real usa 4G/LTE físico
 * - Aqui usamos internet local
 * 
 * O protocolo Modbus e o formato de dados são IDÊNTICOS.
 */

const ModbusRTU = require('modbus-serial');
const axios = require('axios');
const readline = require('readline');

// ═══════════════════════════════════════════════════════════════════
// CONFIGURAÇÃO - Altere conforme necessário
// ═══════════════════════════════════════════════════════════════════

const CONFIG = {
    // Endpoint do NXD (pode ser alterado em runtime)
    nxdEndpoint: process.env.NXD_ENDPOINT || 'https://nxdata-production.up.railway.app/api/ingest',
    
    // API Key (OBRIGATÓRIO - configure antes de iniciar)
    apiKey: process.env.API_KEY || '',
    
    // Intervalo de leitura em ms
    readInterval: 3000,
    
    // Simular latência de 4G (ms)
    simulate4GLatency: true,
    latencyMin: 100,
    latencyMax: 500,
    
    // CLPs para conectar
    clps: [
        {
            name: 'INJETORA_SIEMENS_01',
            brand: 'Siemens',
            model: 'S7-1200',
            host: 'localhost',
            port: 502,
            unitId: 1,
            protocol: 'Modbus TCP'
        },
        {
            name: 'INJETORA_DELTA_01',
            brand: 'Delta',
            model: 'DVP-28SV',
            host: 'localhost',
            port: 503,
            unitId: 2,
            protocol: 'Modbus TCP'
        }
    ]
};

// ═══════════════════════════════════════════════════════════════════
// ESTADO DO GATEWAY
// ═══════════════════════════════════════════════════════════════════

let stats = {
    messagesTotal: 0,
    messagesSuccess: 0,
    messagesError: 0,
    bytesTransmitted: 0,
    startTime: null,
    lastError: null
};

let running = false;
let networkEnabled = true;

// ═══════════════════════════════════════════════════════════════════
// FUNÇÕES DE COMUNICAÇÃO MODBUS
// ═══════════════════════════════════════════════════════════════════

async function readCLP(clpConfig) {
    const client = new ModbusRTU();
    
    try {
        // Conecta ao CLP via Modbus TCP
        await client.connectTCP(clpConfig.host, { port: clpConfig.port });
        client.setID(clpConfig.unitId);
        client.setTimeout(2000);
        
        // Lê Holding Registers 0-9
        const data = await client.readHoldingRegisters(0, 10);
        
        // Fecha conexão
        client.close();
        
        // Decodifica os registros
        const registers = data.data;
        
        return {
            success: true,
            data: {
                device_id: clpConfig.name,
                brand: clpConfig.brand,
                protocol: clpConfig.protocol,
                tags: {
                    Status_Producao: registers[0] === 1,
                    Temperatura_Molde: registers[1] / 10,
                    Pressao_Injecao: registers[2] / 10,
                    Tempo_Ciclo: registers[3] / 10,
                    Total_Pecas: registers[4] + (registers[5] << 16),
                    Consumo_Energia_kWh: registers[6] / 100,
                    Health_Score: registers[7],
                    Alarme_Temperatura: registers[8] === 1,
                    Custo_Hora_Parada: registers[9]
                }
            }
        };
        
    } catch (error) {
        client.close();
        return {
            success: false,
            error: error.message,
            clp: clpConfig.name
        };
    }
}

// ═══════════════════════════════════════════════════════════════════
// FUNÇÕES DE ENVIO HTTP (simula 4G do DX)
// ═══════════════════════════════════════════════════════════════════

async function sendToNXD(data) {
    if (!networkEnabled) {
        throw new Error('Rede desabilitada (simulação de queda)');
    }
    
    // Simula latência de 4G
    if (CONFIG.simulate4GLatency) {
        const latency = CONFIG.latencyMin + Math.random() * (CONFIG.latencyMax - CONFIG.latencyMin);
        await new Promise(resolve => setTimeout(resolve, latency));
    }
    
    const payload = {
        api_key: CONFIG.apiKey,
        ...data
    };
    
    const response = await axios.post(CONFIG.nxdEndpoint, payload, {
        headers: { 'Content-Type': 'application/json' },
        timeout: 10000
    });
    
    return response;
}

// ═══════════════════════════════════════════════════════════════════
// LOOP PRINCIPAL DO GATEWAY
// ═══════════════════════════════════════════════════════════════════

async function gatewayLoop() {
    if (!running) return;
    
    for (const clp of CONFIG.clps) {
        try {
            // 1. Lê dados do CLP via Modbus
            const result = await readCLP(clp);
            
            if (!result.success) {
                console.log(`\n❌ [MODBUS] Erro ao ler ${clp.name}: ${result.error}`);
                stats.messagesError++;
                continue;
            }
            
            // 2. Envia para NXD via HTTP
            stats.messagesTotal++;
            
            const response = await sendToNXD(result.data);
            
            if (response.status === 200) {
                stats.messagesSuccess++;
                stats.bytesTransmitted += JSON.stringify(result.data).length;
                
                const status = result.data.tags.Status_Producao ? '🟢' : '🔴';
                console.log(`\n${status} [${clp.brand}] ${result.data.tags.Total_Pecas} peças | ` +
                           `${result.data.tags.Temperatura_Molde}°C | ` +
                           `${result.data.tags.Health_Score}% saúde → NXD ✓`);
            } else {
                stats.messagesError++;
                console.log(`\n❌ [HTTP] Erro ${response.status} ao enviar ${clp.name}`);
            }
            
        } catch (error) {
            stats.messagesError++;
            stats.lastError = error.message;
            console.log(`\n❌ [GATEWAY] ${clp.name}: ${error.message}`);
        }
    }
    
    // Agenda próxima leitura
    setTimeout(gatewayLoop, CONFIG.readInterval);
}

// ═══════════════════════════════════════════════════════════════════
// INTERFACE DE CONTROLE
// ═══════════════════════════════════════════════════════════════════

const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

function showHelp() {
    console.log(`
┌─────────────────────────────────────────────────────────────────────┐
│                 DELTA DX-2100 GATEWAY SIMULATOR                     │
├─────────────────────────────────────────────────────────────────────┤
│  Comandos:                                                          │
│                                                                     │
│  start    - Inicia o gateway                                       │
│  stop     - Para o gateway                                         │
│  status   - Mostra estatísticas                                    │
│  network  - Toggle queda de rede                                   │
│  config   - Mostra configuração atual                              │
│  api KEY  - Define API Key (ex: api NXD_xxx...)                   │
│  help     - Mostra esta ajuda                                      │
│  exit     - Encerra o programa                                     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
`);
}

function showStatus() {
    const uptime = stats.startTime ? Math.floor((Date.now() - stats.startTime) / 1000) : 0;
    const hours = Math.floor(uptime / 3600);
    const minutes = Math.floor((uptime % 3600) / 60);
    const seconds = uptime % 60;
    
    console.log(`
┌─────────────────────────────────────────────────────────────────────┐
│                      STATUS DO GATEWAY                              │
├─────────────────────────────────────────────────────────────────────┤
│  Estado:           ${running ? '🟢 ATIVO' : '⏸️  PARADO'}                                     │
│  Rede:             ${networkEnabled ? '🌐 OK' : '❌ QUEDA SIMULADA'}                                    │
│  Uptime:           ${hours}h ${minutes}m ${seconds}s                                     │
├─────────────────────────────────────────────────────────────────────┤
│  Mensagens Total:  ${stats.messagesTotal}                                              │
│  Sucesso:          ${stats.messagesSuccess}                                              │
│  Erros:            ${stats.messagesError}                                              │
│  Bytes TX:         ${(stats.bytesTransmitted / 1024).toFixed(2)} KB                                     │
├─────────────────────────────────────────────────────────────────────┤
│  Último Erro:      ${stats.lastError || 'Nenhum'}
└─────────────────────────────────────────────────────────────────────┘
`);
}

function showConfig() {
    console.log(`
┌─────────────────────────────────────────────────────────────────────┐
│                      CONFIGURAÇÃO                                   │
├─────────────────────────────────────────────────────────────────────┤
│  Endpoint:     ${CONFIG.nxdEndpoint}
│  API Key:      ${CONFIG.apiKey ? CONFIG.apiKey.substring(0, 30) + '...' : '❌ NÃO CONFIGURADA'}
│  Intervalo:    ${CONFIG.readInterval}ms
│  Latência 4G:  ${CONFIG.simulate4GLatency ? `${CONFIG.latencyMin}-${CONFIG.latencyMax}ms` : 'Desabilitada'}
├─────────────────────────────────────────────────────────────────────┤
│  CLPs Configurados:                                                 │
`);
    CONFIG.clps.forEach((clp, i) => {
        console.log(`│  ${i+1}. ${clp.name} (${clp.brand}) → ${clp.host}:${clp.port}`);
    });
    console.log(`└─────────────────────────────────────────────────────────────────────┘`);
}

rl.on('line', (input) => {
    const [cmd, ...args] = input.trim().toLowerCase().split(' ');
    
    switch(cmd) {
        case 'start':
            if (!CONFIG.apiKey) {
                console.log('\n❌ Configure a API Key primeiro: api NXD_xxx...');
                break;
            }
            if (running) {
                console.log('\n⚠️  Gateway já está ativo');
                break;
            }
            running = true;
            stats.startTime = Date.now();
            console.log('\n🚀 Gateway INICIADO - Lendo CLPs via Modbus e enviando para NXD');
            gatewayLoop();
            break;
            
        case 'stop':
            running = false;
            console.log('\n⏹️  Gateway PARADO');
            break;
            
        case 'status':
            showStatus();
            break;
            
        case 'network':
            networkEnabled = !networkEnabled;
            console.log(`\n🌐 Rede ${networkEnabled ? 'RESTAURADA' : 'DERRUBADA (simulação)'}`);
            break;
            
        case 'config':
            showConfig();
            break;
            
        case 'api':
            if (args.length > 0) {
                CONFIG.apiKey = args.join(' ').toUpperCase();
                console.log(`\n🔑 API Key configurada: ${CONFIG.apiKey.substring(0, 30)}...`);
            } else {
                console.log('\n❌ Uso: api NXD_sua_chave_aqui');
            }
            break;
            
        case 'help':
            showHelp();
            break;
            
        case 'exit':
        case 'quit':
            console.log('\n👋 Encerrando gateway...');
            process.exit(0);
            break;
            
        default:
            if (cmd) {
                console.log(`\n❓ Comando desconhecido: ${cmd}. Digite 'help' para ver comandos.`);
            }
    }
    
    process.stdout.write('\nDX> ');
});

// ═══════════════════════════════════════════════════════════════════
// INICIALIZAÇÃO
// ═══════════════════════════════════════════════════════════════════

console.log(`
╔═══════════════════════════════════════════════════════════════════════╗
║                                                                       ║
║          ██████╗ ██╗  ██╗      ██████╗  ██╗ ██████╗  ██████╗         ║
║          ██╔══██╗╚██╗██╔╝      ╚════██╗███║██╔═████╗██╔═████╗        ║
║          ██║  ██║ ╚███╔╝  █████╗█████╔╝╚██║██║██╔██║██║██╔██║        ║
║          ██║  ██║ ██╔██╗  ╚════╝██╔═══╝ ██║████╔╝██║████╔╝██║        ║
║          ██████╔╝██╔╝ ██╗       ███████╗██║╚██████╔╝╚██████╔╝        ║
║          ╚═════╝ ╚═╝  ╚═╝       ╚══════╝╚═╝ ╚═════╝  ╚═════╝         ║
║                                                                       ║
║                    INDUSTRIAL IoT GATEWAY SIMULATOR                   ║
║                                                                       ║
╠═══════════════════════════════════════════════════════════════════════╣
║                                                                       ║
║  Este simulador replica EXATAMENTE o comportamento do módulo DX:      ║
║                                                                       ║
║  1. Conecta em CLPs via MODBUS TCP (protocolo industrial real)        ║
║  2. Lê registros Holding Registers                                    ║
║  3. Converte para JSON                                                ║
║  4. Envia via HTTP POST para NXD Cloud                               ║
║  5. Simula latência de rede 4G                                       ║
║                                                                       ║
║  A ÚNICA diferença para o DX real:                                   ║
║  • DX real → Usa chip 4G físico                                      ║
║  • Simulador → Usa sua internet                                       ║
║                                                                       ║
╚═══════════════════════════════════════════════════════════════════════╝

  📌 IMPORTANTE: Antes de iniciar o gateway, você precisa:
     1. Iniciar os CLPs virtuais (clp-siemens.js e clp-delta.js)
     2. Configurar a API Key do NXD

  Digite 'help' para ver os comandos disponíveis.

`);

// Verifica se API Key foi passada por variável de ambiente
if (process.env.API_KEY) {
    CONFIG.apiKey = process.env.API_KEY;
    console.log(`  🔑 API Key carregada da variável de ambiente\n`);
}

process.stdout.write('DX> ');
