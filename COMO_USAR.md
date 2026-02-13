# 🚀 Como Usar o HUB System

## 📋 Pré-requisitos
- Docker Desktop instalado e rodando
- Navegador moderno (Chrome, Firefox, Edge)

---

## ⚡ Início Rápido (3 Passos)

### **Passo 1: Iniciar Servidor**
```bash
# Clique duas vezes em:
START.bat
```

O sistema vai:
- Verificar se Docker está rodando
- Iniciar o servidor HUB
- Abrir o navegador automaticamente

### **Passo 2: Criar Fábrica**
No navegador que abriu (`http://localhost:8080`):
1. Digite o nome da sua fábrica
2. Clique em "Criar Fábrica"
3. **COPIE** a API Key gerada

### **Passo 3: Configurar Simulador**
Crie um arquivo chamado `.env` na raiz do projeto:
```env
API_KEY=cole_sua_chave_aqui
```

Depois execute `START.bat` novamente.

---

## 🎯 O Que Você Vai Ver

Após 5-10 segundos, no dashboard:

### Máquina 1: **Siemens_CLP_SIEMENS_01**
- Temperatura_Motor: 45-65°C
- Pressao_Hidraulica: 120-150 bar
- Velocidade_RPM: 1500-2000
- Status_Producao: true/false
- Contador_Pecas: 0-10000
- Alarme_Temperatura: true/false

### Máquina 2: **Delta_CLP_DELTA_01**
- Temp_Ambiente: 20-35°C
- Corrente_Motor_A: 5-15A
- Tensao_Rede_V: 220-230V
- Ciclos_Completos: 0-5000
- Modo_Operacao: AUTO/MANUAL/SETUP
- Falha_Comunicacao: true/false

**Valores atualizam a cada 2-3 segundos!**

---

## 🐳 Comandos Docker Úteis

```bash
# Ver logs em tempo real
docker-compose logs -f

# Ver logs só do servidor
docker-compose logs -f hub-server

# Ver logs só do simulador
docker-compose logs -f dx-simulator

# Ver status dos containers
docker-compose ps

# Parar tudo
docker-compose down

# Ou simplesmente:
STOP.bat

# Reconstruir containers (após mudanças no código)
docker-compose build --no-cache
docker-compose up -d
```

---

## 🔧 Estrutura de Arquivos

```
HubSystem1.0/
├── START.bat                   ← Clique aqui para iniciar
├── STOP.bat                    ← Clique aqui para parar
├── docker-compose.yml          ← Orquestração dos containers
├── Dockerfile.hub              ← Container do servidor
├── Dockerfile.simulator        ← Container do simulador
├── .env                        ← Suas configurações (criar)
├── .env.example                ← Exemplo de configuração
├── README.md                   ← Documentação principal
├── COMO_USAR.md                ← Este arquivo
│
├── core/                       ← Lógica de negócio
├── api/                        ← Endpoints REST
├── data/                       ← Banco de dados SQLite
├── services/                   ← Serviços (logs, alerts)
├── simulator/                  ← Simulador DX
├── web/                        ← Dashboard HTML/CSS/JS
└── docs/                       ← Documentação técnica
```

---

## 🆘 Problemas Comuns

### ❌ "Docker não está rodando"
**Solução:** Inicie o Docker Desktop e aguarde ficar pronto

### ❌ "Porta 8080 em uso"
**Solução:**
```bash
docker-compose down
# Ou mude a porta em docker-compose.yml
```

### ❌ "Simulador não conecta"
**Solução:**
1. Verifique se a API Key está correta no `.env`
2. Veja os logs: `docker-compose logs dx-simulator`
3. Verifique se o servidor está rodando: `docker-compose ps`

### ❌ "Máquinas não aparecem"
**Solução:**
1. Aguarde 10 segundos
2. Atualize a página (F5)
3. Veja os logs: `docker-compose logs -f`

---

## 📊 Acessos

- **Dashboard:** http://localhost:8080
- **API Health:** http://localhost:8080/api/health
- **Logs:** `docker-compose logs -f`

---

## 🎓 Próximos Passos

1. Teste criar múltiplas fábricas
2. Explore a API REST
3. Veja os logs em tempo real
4. Modifique o código e reconstrua
5. Integre com seu sistema real

---

## 📞 Suporte

Qualquer dúvida ou erro, volte aqui com a mensagem! 🚀
