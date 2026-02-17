# NXD — Decisões e Plano de Implementação

Documento único com as decisões tomadas e o plano para deixar o sistema operando online (Google Cloud, login Google, uma fábrica por assinatura, IA Gemini, termos e políticas).

---

## 1. Decisões confirmadas

### 1.1 API Key
- **Gerada uma vez** na criação da fábrica.
- **Exibição:** mostrar a API Key **só quando o usuário estiver pronto para configurar o DX** (fluxo guiado: “Copie e cole no nanoDX; se perder, gere outra”).
- **Regenerar:** opção na conta para **gerar nova API Key** se perdeu; a antiga deixa de valer; o cliente coloca a nova no nanoDX e religa o circuito.

### 1.2 Assinatura e fábrica
- **Uma assinatura = uma fábrica** (por enquanto).
- Depois pode evoluir para mais de uma fábrica por conta; não é prioridade agora.

### 1.3 Fluxo pós-login
- Usuário **cria login** (Google) → se **não tem fábrica**, tela para **criar fábrica** → se **já tem fábrica**, vai direto para o **dashboard**.

### 1.4 IA (Gemini)
- Usar **Gemini** (já no projeto; período gratuito / limite por plano).
- Relatórios **só técnicos**; **prompts negativos** para evitar conselhos de engenharia ou decisões que não cabem ao sistema.
- **Custo:** limitado por plano (quanto mais uso, mais paga; definido depois).

### 1.5 Dados em tempo real
- NXD **só consulta/recebe**; não “push” do DX no sentido de comando. DX envia dados; NXD pode **atualizar dashboard** em intervalo configurável (ex.: 5 s, 10 s).
- Se houver impacto de performance, **orientar** o usuário a usar intervalo maior (ex.: 10 s em vez de 5 s).

### 1.6 Design e modo demo
- **Funcional primeiro;** design/cosmetics depois.
- **Sem modo demo e sem mock/stub:** sistema 100% operante; testes com **simulador no seu computador** (envio de dados próximo do real). Depois conecta o DX real.

### 1.7 Google Cloud
- Usar e criar **apenas recursos NXD** (prefixo `nxd-*`).
- **Não alterar nem excluir** serviços do SlideFlow.
- Pode ativar APIs, criar recursos novos e configurar o que for necessário para o NXD.

### 1.8 Termos e políticas
- **Termos de Uso** e **Políticas** que deixem claro: NXD **só lê e monitora**; não configura nem controla fábrica; responsabilidade de calibração e manutenção é do cliente; nosso papel é **informar a realidade espelhada** da parte monitorada.
- No **cadastro**, **checkbox obrigatório**: usuário confirma que leu e aceita Termos e Políticas antes de criar conta.

### 1.9 Checklist das 200 perguntas
- Usar as 200 perguntas (Blocos A–D) como **checklist** de segurança e implementação: “Já fizemos isso?” / “Estamos protegidos nisso?”. Não é tudo para a v1; é referência para agora e futuro.

---

## 2. Escopo do primeiro lançamento (o que entra)

- **Autenticação:** login com **Google**; aceite de Termos e Políticas no cadastro.
- **Fábrica:** criar fábrica (uma por conta); API Key gerada uma vez; exibir só no fluxo “pronto para colocar no DX”; opção **regenerar API Key** na conta.
- **Dashboard:** tela principal com **informação geral da fábrica** (cards: máquinas operando, paradas, alertas, produção, etc.).
- **Configurações:** telas de **perfil**, **configuração do sistema**, **FAQ**; links para **Termos de Uso** e **Políticas**.
- **IA:** integração com **Gemini** para **gerar relatórios** (um tipo inicial); relatórios técnicos e prompts negativos.
- **Financeiro:** **um modelo** de tela/relatório financeiro (mínimo viável).
- **Refresh:** intervalo de atualização do dashboard **configurável** (ex.: 5 s, 10 s), com orientação se o servidor sofrer.

---

## 3. O que NÃO entra na v1 (explicitamente)

- Múltiplas fábricas por conta.
- Modo demo ou dados mock.
- Design elaborado (foco em funcional).
- Integração ERP.
- Outras fontes de IA além de Gemini (por enquanto).

---

## 4. Próximos passos de implementação (ordem sugerida)

1. **Documentar** (este arquivo) e manter as 200 perguntas como referência.
2. **Backend:** tabela `users`; fluxo de **Google Auth** (verificar token e criar sessão); vincular **user → factory** (1:1 por enquanto).
3. **Backend:** rotas protegidas por sessão; **regenerar API Key** (invalida antiga, gera nova, retorna uma vez).
4. **Frontend:** tela de **login com Google**; após login, redirecionar para **criar fábrica** ou **dashboard**; fluxo de **exibir API Key uma vez** (“pronto para colocar no DX”) e **regenerar** na conta.
5. **Frontend:** **dashboard** com cards de estado da fábrica; **refresh configurável**.
6. **Frontend:** telas de **configuração** (perfil, sistema, FAQ) e links para **Termos** e **Políticas**.
7. **Backend + Frontend:** **Termos de Uso** e **Políticas** (texto); checkbox no cadastro.
8. **Backend:** endpoint de **relatório com Gemini** (dados factuais do banco + prompts restritivos).
9. **Frontend:** uma tela de **relatório por IA** e **um modelo de tela financeira**.

---

## 5. Variáveis de ambiente (Auth e Google Cloud)

Para login com Google e sessão:

- **GOOGLE_CLIENT_ID** — Client ID do OAuth 2.0 (Google Cloud Console → Credenciais → Criar credencial → ID do cliente OAuth 2.0, tipo “Aplicativo da Web”).
- **GOOGLE_CLIENT_SECRET** — Client secret do mesmo cliente.
- **BASE_URL** — URL base do backend (ex.: `https://nxd-api-xxx.run.app` ou `http://localhost:8080`). Usada como `redirect_uri` do OAuth (deve estar autorizada no cliente).
- **JWT_SECRET** — Segredo para assinar o JWT da sessão (em produção use um valor forte e seguro).
- **FRONTEND_URL** — (opcional) URL do frontend para redirecionar após login; se vazio, usa BASE_URL.

---

## 6. Autorização

Implementar e configurar no Google Cloud (sem mexer no SlideFlow) até o sistema ficar **operando online**: acessar, criar conta (Google), criar fábrica, obter/regenerar API Key, usar dashboard, configurações, IA e uma tela financeira, com Termos e Políticas em vigor.
