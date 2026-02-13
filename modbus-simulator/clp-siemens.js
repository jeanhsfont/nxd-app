/**
 * SIMULADOR CLP SIEMENS S7-1200
 * 
 * Servidor Modbus TCP que simula um CLP Siemens real.
 * Expõe registros Holding Registers que o DX pode ler.
 * 
 * Porta: 502 (padrão Modbus)
 * 
 * Mapa de Registros (Holding Registers):
 * - HR 0: Status Produção (0=Parado, 1=Produzindo)
 * - HR 1: Temperatura Molde (x10, ex: 650 = 65.0°C)
 * - HR 2: Pressão Injeção (x10, ex: 1200 = 120.0 bar)
 * - HR 3: Tempo Ciclo (x10, ex: 450 = 45.0 seg)
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
    port: 502,
    debug: true,
    unitID: 1
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
    1: 650,    // Temperatura: 65.0°C
    2: 1200,   // Pressão: 120.0 bar
    3: 450,    // Tempo Ciclo: 45.0 seg
    4: 0,      // Total Peças (Low)
    5: 0,      // Total Peças (High)
    6: 0,      // Consumo Energia (x100 kWh)
    7: 95,     // Health Score
    8: 0,      // Alarme Temperatura
    9: 850     // Custo Hora: R$ 850
};

// Estado da máquina
let state = {
    running: true,
    totalPecas: 0,
    consumoEnergia: 0,
    temperature: 65,
    healthScore: 95
};

// Simula produção real
function simulateProduction() {
    if (state.running) {
        // Produz peças
        state.totalPecas += Math.floor(Math.random() * 3) + 1;
        
        // Consome energia
        state.consumoEnergia += 0.8 + Math.random() * 0.2;
        
        // Varia temperatura
        state.temperature += (Math.random() - 0.5) * 2;
        state.temperature = Math.max(55, Math.min(90, state.temperature));
        
        // Health score varia levemente
        state.healthScore += (Math.random() - 0.5) * 0.5;
        state.healthScore = Math.max(60, Math.min(100, state.healthScore));
    } else {
        // Máquina parada - temperatura cai
        state.temperature = Math.max(40, state.temperature - 0.5);
    }
    
    // Atualiza registros Modbus
    holdingRegisters[0] = state.running ? 1 : 0;
    holdingRegisters[1] = Math.round(state.temperature * 10);
    holdingRegisters[2] = 1200 + Math.round((Math.random() - 0.5) * 100);
    holdingRegisters[3] = 450 + Math.round((Math.random() - 0.5) * 50);
    holdingRegisters[4] = state.totalPecas & 0xFFFF;           // Low word
    holdingRegisters[5] = (state.totalPecas >> 16) & 0xFFFF;   // High word
    holdingRegisters[6] = Math.round(state.consumoEnergia * 100);
    holdingRegisters[7] = Math.round(state.healthScore);
    holdingRegisters[8] = state.temperature > 85 ? 1 : 0;
}

// Comandos do console
const readline = require('readline');
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

function showStatus() {
    console.log('\n┌─────────────────────────────────────────────────────┐');
    console.log('│          CLP SIEMENS S7-1200 - SIMULADOR            │');
    console.log('├─────────────────────────────────────────────────────┤');
    console.log(`│  Status:      ${state.running ? '🟢 PRODUZINDO' : '🔴 PARADO'}                       │`);
    console.log(`│  Peças:       ${state.totalPecas.toString().padEnd(10)} │`);
    console.log(`│  Temperatura: ${state.temperature.toFixed(1)}°C                            │`);
    console.log(`│  Energia:     ${state.consumoEnergia.toFixed(2)} kWh                       │`);
    console.log(`│  Saúde:       ${state.healthScore.toFixed(0)}%                              │`);
    console.log('├─────────────────────────────────────────────────────┤');
    console.log('│  Comandos: [P]arar  [R]etomar  [S]uperaquecer       │');
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
            state.temperature = 65;
            console.log('🟢 Máquina RETOMOU produção');
            break;
        case 's':
            state.temperature = 95;
            state.running = false;
            state.healthScore = 60;
            console.log('🌡️ SUPERAQUECIMENTO! Máquina parou automaticamente');
            break;
        default:
            showStatus();
    }
});

// Inicia simulação
console.log('═══════════════════════════════════════════════════════════');
console.log('   CLP SIEMENS S7-1200 - SIMULADOR MODBUS TCP');
console.log('   Porta: 502 | Unit ID: 1');
console.log('═══════════════════════════════════════════════════════════');
console.log('');
console.log('   Este servidor simula um CLP Siemens real.');
console.log('   Conecte via Modbus TCP em localhost:502');
console.log('');
console.log('   Comandos: P=Parar | R=Retomar | S=Superaquecer | Enter=Status');
console.log('');

// Atualiza a cada 1 segundo
setInterval(simulateProduction, 1000);

// Mostra status a cada 5 segundos
setInterval(() => {
    if (state.running) {
        process.stdout.write(`\r🟢 Siemens: ${state.totalPecas} peças | ${state.temperature.toFixed(1)}°C | ${state.healthScore.toFixed(0)}% saúde    `);
    } else {
        process.stdout.write(`\r🔴 Siemens: PARADO | ${state.temperature.toFixed(1)}°C                              `);
    }
}, 2000);

server.on('error', (err) => {
    console.error('Erro no servidor Modbus:', err);
});

console.log('✓ Servidor Modbus TCP iniciado na porta 502');
