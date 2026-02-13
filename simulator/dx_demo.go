package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// Configuração da simulação
const (
	// 10 minutos reais = 10 horas de fábrica (8h às 18h)
	SIMULATION_DURATION = 10 * time.Minute
	FACTORY_HOURS       = 10.0 // 8h às 18h = 10 horas
	TIME_SCALE          = FACTORY_HOURS * 60 / 10 // 1 segundo real = 1 minuto de fábrica

	// Custos por hora parada
	CUSTO_SIEMENS = 850.0 // R$/hora
	CUSTO_DELTA   = 650.0 // R$/hora

	// Produção por ciclo
	PECAS_POR_CICLO_SIEMENS = 12
	PECAS_POR_CICLO_DELTA   = 10
	TEMPO_CICLO_BASE        = 45.0 // segundos (tempo de fábrica)
)

// Estrutura de dados para envio
type DataPayload struct {
	APIKey   string                 `json:"api_key"`
	DeviceID string                 `json:"device_id"`
	Brand    string                 `json:"brand"`
	Protocol string                 `json:"protocol"`
	Tags     map[string]interface{} `json:"tags"`
}

// Estado de cada máquina
type MachineState struct {
	Name              string
	Brand             string
	Model             string
	CustoHora         float64
	PecasPorCiclo     int
	Running           bool
	Temperature       float64
	Pressure          float64
	CycleTime         float64
	TotalPecas        int
	ConsumoEnergia    float64
	TempoParada       float64 // minutos de fábrica
	UltimaParada      time.Time
	HealthScore       float64 // 0-100
	TempHistory       []float64
}

// Eventos programados (com variação)
type ScheduledEvent struct {
	FactoryMinute float64 // minuto do dia de fábrica (0-600 para 10h)
	MachineID     string
	EventType     string // "stop", "start", "overheat", "comm_fail"
	Duration      float64 // duração em minutos de fábrica
	Triggered     bool
}

var (
	endpoint string
	apiKey   string
	machines map[string]*MachineState
	events   []ScheduledEvent
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	log.Println("🏭 ══════════════════════════════════════════════════════════")
	log.Println("🏭  SIMULADOR NXD - Demo Fábrica Vale Plast")
	log.Println("🏭  Produto: Tampas plásticas para garrafas PET")
	log.Println("🏭  Simulação: 1 dia de produção (8h-18h) em 10 minutos")
	log.Println("🏭 ══════════════════════════════════════════════════════════")

	// Configuração
	endpoint = os.Getenv("NXD_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8080/api/ingest"
	}

	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("❌ API_KEY não configurada")
	}

	log.Printf("📡 Endpoint: %s", endpoint)
	log.Printf("🔑 API Key: %s...", apiKey[:20])

	// Inicializa máquinas
	initMachines()

	// Programa eventos do dia (com variação aleatória)
	scheduleEvents()

	log.Println("")
	log.Println("⏰ Iniciando simulação do dia de produção...")
	log.Println("   08:00 - Início do turno")
	log.Println("")

	// Loop principal de simulação
	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime)
			
			// Calcula o "horário de fábrica"
			factoryMinutes := elapsed.Seconds() * TIME_SCALE
			factoryHour := 8 + int(factoryMinutes/60)
			factoryMin := int(factoryMinutes) % 60

			// Verifica se simulação terminou
			if elapsed >= SIMULATION_DURATION {
				printFinalReport(factoryMinutes)
				return
			}

			// Processa eventos programados
			processEvents(factoryMinutes)

			// Atualiza e envia dados das máquinas
			for id, machine := range machines {
				updateMachine(machine, factoryMinutes)
				sendData(id, machine, factoryHour, factoryMin)
			}
		}
	}
}

func initMachines() {
	machines = make(map[string]*MachineState)

	machines["INJETORA_SIEMENS_01"] = &MachineState{
		Name:          "Injetora_Siemens_01",
		Brand:         "Siemens",
		Model:         "S7-1200",
		CustoHora:     CUSTO_SIEMENS,
		PecasPorCiclo: PECAS_POR_CICLO_SIEMENS,
		Running:       true,
		Temperature:   65.0,
		Pressure:      120.0,
		CycleTime:     TEMPO_CICLO_BASE,
		HealthScore:   95.0,
		TempHistory:   make([]float64, 0),
	}

	machines["INJETORA_DELTA_01"] = &MachineState{
		Name:          "Injetora_Delta_01",
		Brand:         "Delta",
		Model:         "DVP-28SV",
		CustoHora:     CUSTO_DELTA,
		PecasPorCiclo: PECAS_POR_CICLO_DELTA,
		Running:       true,
		Temperature:   62.0,
		Pressure:      115.0,
		CycleTime:     TEMPO_CICLO_BASE + 5,
		HealthScore:   92.0,
		TempHistory:   make([]float64, 0),
	}

	log.Println("✓ Máquinas inicializadas:")
	log.Printf("  • Siemens S7-1200 (Custo parada: R$ %.0f/hora)", CUSTO_SIEMENS)
	log.Printf("  • Delta DVP-28SV (Custo parada: R$ %.0f/hora)", CUSTO_DELTA)
}

func scheduleEvents() {
	events = []ScheduledEvent{}

	// Evento 1: Siemens superaquece (~10:30, variação ±15min)
	// 10:30 = 150 minutos desde 8:00
	events = append(events, ScheduledEvent{
		FactoryMinute: 150 + float64(rand.Intn(30)-15),
		MachineID:     "INJETORA_SIEMENS_01",
		EventType:     "overheat",
		Duration:      45 + float64(rand.Intn(20)-10), // 35-55 min parada
	})

	// Evento 2: Almoço (12:00-13:00)
	// 12:00 = 240 minutos desde 8:00
	events = append(events, ScheduledEvent{
		FactoryMinute: 240,
		MachineID:     "ALL",
		EventType:     "lunch_start",
		Duration:      60,
	})

	events = append(events, ScheduledEvent{
		FactoryMinute: 300, // 13:00
		MachineID:     "ALL",
		EventType:     "lunch_end",
	})

	// Evento 3: Delta falha de comunicação (~16:00, variação ±10min)
	// 16:00 = 480 minutos desde 8:00
	events = append(events, ScheduledEvent{
		FactoryMinute: 480 + float64(rand.Intn(20)-10),
		MachineID:     "INJETORA_DELTA_01",
		EventType:     "comm_fail",
		Duration:      20 + float64(rand.Intn(10)-5), // 15-25 min parada
	})

	log.Println("✓ Eventos do dia programados (com variação aleatória)")
}

func processEvents(factoryMinutes float64) {
	for i := range events {
		if events[i].Triggered {
			continue
		}

		if factoryMinutes >= events[i].FactoryMinute {
			events[i].Triggered = true
			triggerEvent(&events[i], factoryMinutes)
		}
	}

	// Verifica fim de eventos de parada
	for _, machine := range machines {
		if !machine.Running && machine.UltimaParada.Add(time.Duration(machine.TempoParada/TIME_SCALE)*time.Second).Before(time.Now()) {
			machine.Running = true
			factoryHour := 8 + int(factoryMinutes/60)
			factoryMin := int(factoryMinutes) % 60
			log.Printf("🟢 [%02d:%02d] %s VOLTOU a produzir", factoryHour, factoryMin, machine.Name)
		}
	}
}

func triggerEvent(event *ScheduledEvent, factoryMinutes float64) {
	factoryHour := 8 + int(factoryMinutes/60)
	factoryMin := int(factoryMinutes) % 60

	switch event.EventType {
	case "overheat":
		machine := machines[event.MachineID]
		machine.Running = false
		machine.Temperature = 95.0 + rand.Float64()*10 // Temperatura alta
		machine.TempoParada = event.Duration
		machine.UltimaParada = time.Now()
		machine.HealthScore -= 15
		log.Printf("🔴 [%02d:%02d] ⚠️  %s PAROU - SUPERAQUECIMENTO! (Temp: %.1f°C)", 
			factoryHour, factoryMin, machine.Name, machine.Temperature)
		log.Printf("   💸 Prejuízo estimado: R$ %.0f (%.0f min × R$ %.0f/hora)",
			event.Duration/60*machine.CustoHora, event.Duration, machine.CustoHora)

	case "comm_fail":
		machine := machines[event.MachineID]
		machine.Running = false
		machine.TempoParada = event.Duration
		machine.UltimaParada = time.Now()
		machine.HealthScore -= 8
		log.Printf("🔴 [%02d:%02d] ⚠️  %s PAROU - FALHA DE COMUNICAÇÃO!", 
			factoryHour, factoryMin, machine.Name)
		log.Printf("   💸 Prejuízo estimado: R$ %.0f (%.0f min × R$ %.0f/hora)",
			event.Duration/60*machine.CustoHora, event.Duration, machine.CustoHora)

	case "lunch_start":
		log.Printf("🍽️  [%02d:%02d] INTERVALO DE ALMOÇO - Máquinas em standby", factoryHour, factoryMin)
		for _, machine := range machines {
			machine.Running = false
			machine.TempoParada = event.Duration
			machine.UltimaParada = time.Now()
		}

	case "lunch_end":
		log.Printf("🟢 [%02d:%02d] FIM DO ALMOÇO - Retomando produção", factoryHour, factoryMin)
		for _, machine := range machines {
			machine.Running = true
		}
	}
}

func updateMachine(machine *MachineState, factoryMinutes float64) {
	if machine.Running {
		// Produção normal
		machine.TotalPecas += machine.PecasPorCiclo
		machine.ConsumoEnergia += 2.5 + rand.Float64()*0.5

		// Temperatura oscila naturalmente
		tempVariation := (rand.Float64() - 0.5) * 4
		machine.Temperature = clamp(machine.Temperature+tempVariation, 55, 85)

		// Pressão oscila
		pressVariation := (rand.Float64() - 0.5) * 6
		machine.Pressure = clamp(machine.Pressure+pressVariation, 100, 140)

		// Tempo de ciclo varia um pouco
		machine.CycleTime = TEMPO_CICLO_BASE + (rand.Float64()-0.5)*5

		// Health score recupera lentamente
		if machine.HealthScore < 95 {
			machine.HealthScore += 0.1
		}
	} else {
		// Máquina parada - temperatura cai aos poucos
		if machine.Temperature > 40 {
			machine.Temperature -= 0.5
		}
		// Energia mínima em standby
		machine.ConsumoEnergia += 0.3
	}

	// Guarda histórico de temperatura para análise
	machine.TempHistory = append(machine.TempHistory, machine.Temperature)
	if len(machine.TempHistory) > 30 {
		machine.TempHistory = machine.TempHistory[1:]
	}

	// Calcula variação de temperatura (para indicador de saúde)
	if len(machine.TempHistory) > 5 {
		variance := calculateVariance(machine.TempHistory)
		if variance > 20 {
			machine.HealthScore = clamp(machine.HealthScore-0.5, 0, 100)
		}
	}
}

func sendData(id string, machine *MachineState, hour, min int) {
	payload := DataPayload{
		APIKey:   apiKey,
		DeviceID: id,
		Brand:    machine.Brand,
		Protocol: "Modbus",
		Tags: map[string]interface{}{
			"Status_Producao":    machine.Running,
			"Temperatura_Molde":  math.Round(machine.Temperature*10) / 10,
			"Pressao_Injecao":    math.Round(machine.Pressure*10) / 10,
			"Tempo_Ciclo":        math.Round(machine.CycleTime*10) / 10,
			"Total_Pecas":        machine.TotalPecas,
			"Consumo_Energia_kWh": math.Round(machine.ConsumoEnergia*100) / 100,
			"Health_Score":       math.Round(machine.HealthScore*10) / 10,
			"Alarme_Temperatura": machine.Temperature > 85,
			"Custo_Hora_Parada":  machine.CustoHora,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ Erro ao serializar dados: %v", err)
		return
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("⚠️  [%02d:%02d] %s - Erro de conexão: %v", hour, min, machine.Name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		status := "🟢"
		if !machine.Running {
			status = "🔴"
		}
		// Log simplificado a cada envio
		if rand.Float64() < 0.3 { // Só mostra 30% dos logs para não poluir
			log.Printf("%s [%02d:%02d] %s | Peças: %d | Temp: %.1f°C | Saúde: %.0f%%",
				status, hour, min, machine.Brand, machine.TotalPecas, machine.Temperature, machine.HealthScore)
		}
	} else {
		log.Printf("❌ [%02d:%02d] %s - Status: %d", hour, min, machine.Name, resp.StatusCode)
	}
}

func printFinalReport(totalMinutes float64) {
	log.Println("")
	log.Println("🏁 ══════════════════════════════════════════════════════════")
	log.Println("🏁  FIM DO TURNO - RELATÓRIO DO DIA")
	log.Println("🏁 ══════════════════════════════════════════════════════════")
	log.Println("")

	siemens := machines["INJETORA_SIEMENS_01"]
	delta := machines["INJETORA_DELTA_01"]

	log.Println("📊 PRODUÇÃO:")
	log.Printf("   Siemens: %d peças", siemens.TotalPecas)
	log.Printf("   Delta:   %d peças", delta.TotalPecas)
	log.Printf("   Total:   %d peças", siemens.TotalPecas+delta.TotalPecas)
	log.Println("")

	log.Println("⚡ CONSUMO DE ENERGIA:")
	log.Printf("   Siemens: %.1f kWh", siemens.ConsumoEnergia)
	log.Printf("   Delta:   %.1f kWh", delta.ConsumoEnergia)
	log.Println("")

	log.Println("📈 EFICIÊNCIA (Peças/kWh):")
	efSiemens := float64(siemens.TotalPecas) / siemens.ConsumoEnergia
	efDelta := float64(delta.TotalPecas) / delta.ConsumoEnergia
	log.Printf("   Siemens: %.1f peças/kWh", efSiemens)
	log.Printf("   Delta:   %.1f peças/kWh", efDelta)
	if efSiemens > efDelta {
		log.Printf("   ✓ Siemens %.1f%% mais eficiente", ((efSiemens/efDelta)-1)*100)
	} else {
		log.Printf("   ✓ Delta %.1f%% mais eficiente", ((efDelta/efSiemens)-1)*100)
	}
	log.Println("")

	log.Println("💚 SAÚDE DAS MÁQUINAS:")
	log.Printf("   Siemens: %.0f%%", siemens.HealthScore)
	log.Printf("   Delta:   %.0f%%", delta.HealthScore)
	log.Println("")

	log.Println("🏭 Simulação concluída! Dados enviados para o NXD.")
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func calculateVariance(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	mean := 0.0
	for _, v := range data {
		mean += v
	}
	mean /= float64(len(data))

	variance := 0.0
	for _, v := range data {
		variance += (v - mean) * (v - mean)
	}
	return variance / float64(len(data))
}
