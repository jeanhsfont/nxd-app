/**
 * SIMULADOR CLP DELTA DVP-28SV
 * 
 * Servidor Modbus TCP que simula um CLP Delta real.
 * Expõe registros Holding Registers que o DX pode ler.
 * 
 * Porta: 503 (diferente do Siemens)
 * 
 * Mapa de Registros (Holding Registers):
 * - HR 0: Status Produção (0=Parado, 1=Produzindo)
 * - HR 1: Temperatura Molde (x10, ex: 580 = 58.0°C)
 * - HR 2: Pressão Injeção (x10, ex: 1150 = 115.0 bar)
 * - HR 3: Tempo Ciclo (x10, ex: 480 = 48.0 seg)
 * - HR 4-5: Total Peças (32 bits, HR4=Low, HR5=High)
 * - HR 6: Consumo Energia (x100, kWh)
 * - HR 7: Health Score (0-100)
 * - HR 8: Alarme Temperatura (0=Normal, 1=Alarme)
 * - HR 9: Custo Hora Parada (R$/hora)
 */

const ModbusRTU = require('modbus-serial');

// Cria servidor Modbus TCP
const server = new ModbusRTU.ServerTCP({
    host: '0.0.0.0',
    port: 503,
    debug: true,
    unitID: 2
}, {
    // Callback para leitura de Holding Registers
    getHoldingRegister: function(addr, unitID, callback) {
        callback(null, holdingRegisters[addr] || 0);
    },
    // Callback para múltiplos registros
    getMultipleHoldingRegisters: function(startAddr, length, unitID, callback) {
        const values = [];
        for (let i = 0; i < length; i++) {
            values.push(holdingRegisters[startAddr + i] || 0);
        }
        callback(null, values);
    }
});

// Registros do CLP (iniciais)
let holdingRegisters = {
    0: 1,      // Status: Produzindo
    1: 580,    // Temperatura: 58.0°C
    2: 1150,   // Pressão: 115.0 bar
    3: 480,    // Tempo Ciclo: 48.0 seg
    4: 0,      // Total Peças (Low)
    5: 0,      // Total Peças (High)
    6: 0,      // Consumo Energia (x100 kWh)
    7: 92,     // Health Score
    8: 0,      // Alarme Temperatura
    9: 650     // Custo Hora: R$ 650
};

// Estado da máquina
let state = {
    running: true,
    totalPecas: 0,
    consumoEnergia: 0,
    temperature: 58,
    healthScore: 92,
    commError: false
};

// Simula produção real
function simulateProduction() {
    if (state.running && !state.commError) {
        // Produz peças (Delta um pouco mais lenta)
        state.totalPecas += Math.floor(Math.random() * 2) + 1;
        
        // Consome energia (Delta mais eficiente)
        state.consumoEnergia += 0.6 + Math.random() * 0.2;
        
        // Varia temperatura (Delta mais fria)
        state.temperature += (Math.random() - 0.5) * 1.5;
        state.temperature = Math.max(50, Math.min(85, state.temperature));
        
        // Health score
        state.healthScore += (Math.random() - 0.5) * 0.3;
        state.healthScore = Math.max(70, Math.min(100, state.healthScore));
    } else {
        // Máquina parada - temperatura cai
        state.temperature = Math.max(35, state.temperature - 0.3);
    }
    
    // Atualiza registros Modbus
    holdingRegisters[0] = state.running ? 1 : 0;
    holdingRegisters[1] = Math.round(state.temperature * 10);
    holdingRegisters[2] = 1150 + Math.round((Math.random() - 0.5) * 80);
    holdingRegisters[3] = 480 + Math.round((Math.random() - 0.5) * 40);
    holdingRegisters[4] = state.totalPecas & 0xFFFF;           // Low word
    holdingRegisters[5] = (state.totalPecas >> 16) & 0xFFFF;   // High word
    holdingRegisters[6] = Math.round(state.consumoEnergia * 100);
    holdingRegisters[7] = Math.round(state.healthScore);
    holdingRegisters[8] = state.temperature > 80 ? 1 : 0;
}

// Comandos do console
const readline = require('readline');
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

function showStatus() {
    console.log('\n┌─────────────────────────────────────────────────────┐');
    console.log('│           CLP DELTA DVP-28SV - SIMULADOR            │');
    console.log('├─────────────────────────────────────────────────────┤');
    console.log(`│  Status:      ${state.running ? '🟢 PRODUZINDO' : '🔴 PARADO'}                       │`);
    console.log(`│  Peças:       ${state.totalPecas.toString().padEnd(10)} │`);
    console.log(`│  Temperatura: ${state.temperature.toFixed(1)}°C                            │`);
    console.log(`│  Energia:     ${state.consumoEnergia.toFixed(2)} kWh                       │`);
    console.log(`│  Saúde:       ${state.healthScore.toFixed(0)}%                              │`);
    console.log('├─────────────────────────────────────────────────────┤');
    console.log('│  Comandos: [P]arar  [R]etomar  [C]omm Fail          │');
    console.log('└─────────────────────────────────────────────────────┘');
}

rl.on('line', (input) => {
    const cmd = input.toLowerCase().trim();
    switch(cmd) {
        case 'p':
            state.running = false;
            console.log('🔴 Máquina PARADA');
            break;
        case 'r':
            state.running = true;
            state.commError = false;
            state.temperature = 58;
            console.log('🟢 Máquina RETOMOU produção');
            break;
        case 'c':
            state.commError = true;
            state.running = false;
            console.log('📡 FALHA DE COMUNICAÇÃO! Máquina offline');
            break;
        default:
            showStatus();
    }
});

// Inicia simulação
console.log('═══════════════════════════════════════════════════════════');
console.log('   CLP DELTA DVP-28SV - SIMULADOR MODBUS TCP');
console.log('   Porta: 503 | Unit ID: 2');
console.log('═══════════════════════════════════════════════════════════');
console.log('');
console.log('   Este servidor simula um CLP Delta real.');
console.log('   Conecte via Modbus TCP em localhost:503');
console.log('');
console.log('   Comandos: P=Parar | R=Retomar | C=Comm Fail | Enter=Status');
console.log('');

// Atualiza a cada 1 segundo
setInterval(simulateProduction, 1000);

// Mostra status a cada 5 segundos
setInterval(() => {
    if (state.running) {
        process.stdout.write(`\r🟢 Delta: ${state.totalPecas} peças | ${state.temperature.toFixed(1)}°C | ${state.healthScore.toFixed(0)}% saúde      `);
    } else if (state.commError) {
        process.stdout.write(`\r📡 Delta: FALHA COMUNICAÇÃO                                  `);
    } else {
        process.stdout.write(`\r🔴 Delta: PARADO | ${state.temperature.toFixed(1)}°C                                `);
    }
}, 2000);

server.on('error', (err) => {
    console.error('Erro no servidor Modbus:', err);
});

console.log('✓ Servidor Modbus TCP iniciado na porta 503');
