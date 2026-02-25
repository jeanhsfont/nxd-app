# Análise NXD — UX, UI, Engenharia e Teste (Ótica ISO / Apresentável)

**Data:** Fevereiro 2026  
**Escopo:** Sistema NXD (Nexus Data Exchange) — proposta de valor, fluxos, interface, navegação, anomalias e recomendações para impressionar em demonstração e venda.

---

## 1. Proposta do NXD (referência para a nota)

- **O que é:** Plataforma de monitoramento e troca de dados industriais; conecta máquinas/CLPs a um painel centralizado; ingestão em tempo real, histórico, relatórios, indicadores financeiros e IA (Gemini).
- **Quem usa:** Fábricas (uma por assinatura na v1); usuário cria conta → onboarding → API Key → dashboard com CLPs/ativos.
- **Diferencial:** Dados reais (sem mock), IA com telemetria, foco em “informar a realidade espelhada” da planta, sem controle remoto.

---

## 2. Nota geral (0–10) e resumo

| Critério | Nota (0–10) | Comentário breve |
|----------|-------------|-------------------|
| **Proposta de valor (fit com o produto)** | 8 | Dashboard escuro, métricas claras, IA e financeiro alinhados ao NXD. Falta “wow” na primeira tela (hero/valor). |
| **Consistência visual e identidade** | 6 | Login/Register/Onboarding genéricos (cinza/azul); Dashboard escuro e moderno; sidebar indigo. Dois mundos visuais. |
| **Navegação e arquitetura de informação** | 6 | Sidebar boa; rota `/dashboard` solta; Termos/Suporte saem do layout; sem 404; páginas órfãs (Sectors, IAReports). |
| **Fluxo crítico (registro → primeiro valor)** | 6 | Register → Login → Onboarding funciona; falta aceite de Termos no cadastro; Onboarding passo 3 com texto errado; API Key bem explicada. |
| **Estados (loading, vazio, erro)** | 7 | Dashboard: FirstLoad, skeleton, vazio; Billing/Settings/Import com loading; vários erros tratados. FirstLoad é simulado (tempo), não progresso real. |
| **Acessibilidade (WCAG / uso por todos)** | 5 | Vários labels ok; falta aria em modais e feedback de erro estruturado; foco em modais e teclado não validados. |
| **Confiabilidade e robustez** | 6 | Auth e API corrigidos; falta 404, tratamento de token expirado e retry em falhas de rede. |
| **Apresentabilidade (demo/venda)** | 6 | Funcional e credível; precisa de polish (copy, telas de boas-vindas, 404, termos no registro). |

**Nota final proposta: 6,5 / 10** — Base sólida e funcional; falta alinhar identidade, corrigir anomalias críticas e adicionar detalhes que elevem a “pronto para impressionar”.

---

## 3. Anomalias críticas (corrigir antes de apresentar)

### 3.1 Conteúdo e copy

- **Onboarding — Passo 3:** Título é “Segurança da Conta”, mas o texto diz: *“Configure os dados da sua fábrica e gere sua chave de API…”*. Esse texto é do passo 2; no passo 3 deveria falar de 2FA e, em seguida, “Finalizar e gerar chave”.
- **ApiKeyModal:** Aviso contém **markdown** (`**única vez**`) que não é renderizado; fica literal no texto. Trocar por `<strong>` ou texto sem markdown.
- **Register:** Documento NXD exige *checkbox obrigatório de aceite dos Termos e Políticas* antes de criar conta. Hoje não existe; link para Termos existe, mas não há “Li e aceito os Termos”.

### 3.2 Navegação e rotas

- **Rota `/dashboard`:** Está **fora** de `ProtectedRoute` e **sem** `MainLayout`. Acesso direto a `/dashboard` mostra o dashboard sem sidebar e, dependendo do código, pode quebrar ou expor dados. Deve ser protegida e, se for alias de “início”, redirecionar para `/` ou usar o mesmo layout.
- **Página 404:** Não existe rota catch-all (`path="*"`). URL inexistente mostra tela em branco ou comportamento estranho. Necessário componente “Página não encontrada” com link para início/login.
- **Termos e Suporte na sidebar:** Ao clicar em “Termos” ou “Suporte”, o usuário **sai** do layout (nova rota sem sidebar). Perde contexto e “saiu do app”. Melhor: abrir em nova aba ou manter sidebar com conteúdo de Termos/Suporte no conteúdo principal.

### 3.3 Segurança e fluxo

- **ProtectedRoute:** Só verifica existência do token no `localStorage`; não valida expiração nem renovação. Token expirado pode gerar 401 em toda requisição sem mensagem clara (“Sessão expirada, faça login novamente”) nem redireciono automático para login.
- **Support:** POST `/api/support` exige autenticação. Se o usuário acessar `/support` sem estar logado, o envio falha sem explicação amigável (ex.: “Faça login para enviar mensagem”).

### 3.4 UX de formulários e feedback

- **Login/Register:** Mensagens de erro aparecem em texto pequeno e itálico; não há indicação de campo inválido (ex.: borda vermelha no input). Quem usa leitor de tela não tem `aria-live` para erros.
- **Onboarding — Finalizar:** Em caso de falha no POST (rede, 500), o usuário só vê `console.error`; não há `setError` ou toast. Deve haver mensagem visível e opção de tentar novamente.

---

## 4. Anomalias médias (recomendado corrigir)

- **FirstLoadModal (Dashboard):** Passos e barra de progresso são **tempo-dependentes** (0.5s, 1.5s, 3s…), não refletem o progresso real da API. Se a rede for lenta, o usuário vê “100%” antes dos dados chegarem. Preferível: progresso atrelado a etapas reais (ex.: “dados recebidos” → próximo passo) ou um único spinner com “Carregando…”.
- **Páginas sem rota:** `Sectors.jsx` e `IAReports.jsx` não estão em `App.jsx`. Ou são integradas (ex.: Sectors dentro de Assets, IA Reports em IA) ou removidas/arquivadas para não gerar confusão.
- **Billing:** Mensagem técnica “Configure STRIPE_CHECKOUT_URL…” visível para o usuário final quando não há checkout. Deve ser mensagem amigável: “Pagamento em breve. Entre em contato para upgrade.”
- **Consistência de cores:** Register usa botão verde; Login usa azul; Onboarding mistura azul/cinza/verde. Padronizar primária (ex.: indigo do layout) em todos os fluxos de auth.
- **Sair (logout):** Apenas remove token e redireciona. Não há confirmação (“Deseja mesmo sair?”). Opcional, mas comum em apps corporativos.

---

## 5. Pontos fortes (manter e destacar)

- **Dashboard:** Visual escuro, cards de métricas, barras de status, Health Score, “Produzindo/Parado”, último visto — transmite “operacional” e confiança.
- **ApiKeyModal:** Aviso forte, checkbox de confirmação, botão “Copiar”, desabilitar “Concluir” até marcar — bom para segurança e compliance.
- **Onboarding em passos:** Dados pessoais → fábrica (CNPJ com busca) → 2FA → chave; fluxo claro.
- **Busca por CNPJ:** Pré-preenchimento com dados públicos agiliza cadastro e reduz erro.
- **NXD Intelligence (Chat):** Boas-vindas claras, setor opcional, indicador “Analisando telemetria” — diferencia o produto.
- **Termos de Uso:** Texto completo, seções numeradas, data de atualização — adequado para jurídico e apresentação.
- **Estados vazios:** Dashboard (“Nenhum CLP detectado”), Import (“Nenhum job”), Financial e Assets com mensagens e próximos passos.

---

## 6. Ideias para impressionar (além do óbvio)

### 6.1 Primeira impressão e “wow”

- **Landing / primeira tela após login:** Se o usuário já tem fábrica, em vez de ir direto ao dashboard “cru”, mostrar uma **tela de boas-vindas** curta: “Olá, [Nome]. Sua fábrica [Nome Fábrica] está conectada.” + um único KPI (ex.: “X CLPs online”) e botão “Ver dashboard”. Cria momento de valor em 5 segundos.
- **Dashboard vazio:** Em vez de só “Nenhum CLP detectado” + comando do simulador, adicionar um **mini-tour** ou um passo a passo visual (1. Obtenha sua API Key em Ajustes 2. Configure o DX Simulator 3. Veja os dados aqui) com links. Reduz tempo até o “aha moment”.
- **Número de “dias conectado” ou “primeira conexão”:** Ex.: “Conectado há 12 dias” ou “Primeira telemetria recebida em DD/MM”. Gera sensação de histórico e confiança.

### 6.2 Confiança e credibilidade

- **Indicador de “última sincronia do sistema”:** No header do dashboard, algo como “Dados atualizados há X segundos” (já existe “Atualizado X atrás”; garantir que seja consistente com o refresh real).
- **Modo “apresentação”:** Botão ou query param que oculta dados sensíveis (nomes de fábrica reais, valores absolutos) e mostra dados de exemplo ou genéricos para demo em tela grande.
- **Página “Status do serviço” (opcional):** Link no rodapé ou Suporte para uma página estática ou status page (ex.: “Todos os sistemas operacionais”) para transmitir profissionalismo.

### 6.3 Industrial e operação (referência IIoT)

- **Limitar métricas em destaque:** Literatura sugere 5–7 métricas por vista para não sobrecarregar. O dashboard já foca em total/online/offline/disponibilidade + cards por ativo; evitar adicionar dezenas de KPIs na mesma tela.
- **Refresh configurável:** Documento NXD cita “intervalo de atualização configurável (5s, 10s)”. Se já existir no backend, expor no front (dropdown ou slider) e aviso se “intervalo muito curto pode impactar performance”.
- **Cores semânticas consistentes:** Verde = ok/online, âmbar = atenção, vermelho = falha. Dashboard já faz isso; manter em todas as telas (Billing, Settings, Import).

### 6.4 Onboarding (referência SaaS)

- **Barra de progresso no onboarding:** Mostrar “Passo 1 de 3”, “Passo 2 de 3” (e opcionalmente %). Aumenta conclusão.
- **Uma pergunta de “objetivo” no início:** Ex.: “O que você quer fazer primeiro? Conectar CLPs / Ver relatórios / Configurar cobrança.” Não precisa mudar o fluxo; só permite personalizar a mensagem após o onboarding (“Você escolheu X; comece por aqui”).
- **E-mail pós-cadastro:** Após gerar a API Key, enviar e-mail com “Sua chave foi gerada” + link para Ajustes e para o simulador (se tiver integração de e-mail).

### 6.5 Acessibilidade e inclusão

- **Labels em todos os inputs:** Garantir que todo `<input>` tenha `<label htmlFor="...">` (Register/Login já têm em parte; revisar Onboarding e Support).
- **Erros associados aos campos:** Usar `aria-describedby` no input que falhou, apontando para o id do texto de erro, e `aria-invalid="true"` quando houver validação.
- **Foco em modais:** Ao abrir ApiKeyModal ou RegenerateKeyModal, mover foco para o primeiro elemento interativo e manter tab dentro do modal (focus trap); ao fechar, devolver foco ao botão que abriu.

### 6.6 Pequenos polish

- **Favicon e título:** Verificar se há favicon e `<title>` por rota (ex.: “Dashboard | NXD”, “Login | NXD”) para abas e bookmarks.
- **Logo na tela de login:** Reutilizar o mesmo ícone/texto “NXD v2.0” da sidebar para reforçar identidade.
- **Link “Esqueci minha senha”:** Na tela de login, mesmo que não implementado ainda, colocar o link com mensagem “Em breve” ou “Contate o suporte” para mostrar que foi pensado.

---

## 7. Checklist de ações (priorizado)

### Crítico (antes de apresentar)

1. Corrigir **texto do Onboarding passo 3** (Segurança da Conta vs. dados da fábrica).
2. Remover **markdown** do aviso no **ApiKeyModal** (ou renderizar HTML).
3. Incluir **checkbox “Li e aceito os Termos de Uso”** no **Register**, obrigatório, com link para `/terms`.
4. Proteger **rota `/dashboard`** (ex.: redirecionar para `/` ou colocar dentro de ProtectedRoute + MainLayout).
5. Criar **página 404** e rota `path="*"` com mensagem e link para `/` ou `/login`.
6. Tratar **token expirado** (401): mensagem “Sessão expirada” e redireciono para `/login`; opcionalmente limpar token.

### Recomendado (curto prazo)

7. **FirstLoadModal:** Atrelar a progresso real ou substituir por spinner + “Carregando…”.
8. **Termos/Suporte:** Manter no layout (conteúdo no `<Outlet />`) ou abrir em nova aba.
9. **Support sem login:** Mensagem clara “Faça login para enviar mensagem” e link para login.
10. **Onboarding submit:** Mostrar erro na tela em caso de falha e botão “Tentar novamente”.
11. **Billing:** Trocar mensagem técnica por texto amigável quando não houver checkout.
12. **Cores de auth:** Unificar (ex.: indigo) em Login, Register e Onboarding.

### Desejável (polish)

13. Barra de progresso no Onboarding (Passo X de 3).
14. Tela de boas-vindas pós-login (uma frase + um KPI).
15. Dashboard vazio com mini-tour ou passos visuais.
16. Acessibilidade: `aria-describedby`/`aria-invalid` em erros; focus trap em modais.
17. Favicon e títulos por página.
18. Link “Esqueci minha senha” no Login (mesmo que “Em breve”).

---

## 8. Referências utilizadas

- Documento interno: `NXD_DECISOES_E_PLANO.md`
- Mapa de rotas e fluxos: exploração do repositório (App.jsx, MainLayout, páginas em `web-app/src`).
- UX IIoT: best practices para dashboards industriais (métricas focadas, operador em primeiro, credibilidade dos dados).
- SaaS onboarding: simplificação, progresso visível, tempo até valor, personalização.
- Acessibilidade: WCAG, labels em formulários, gestão de foco.

---

**Conclusão:** O NXD está **funcional e credível** para demonstração e venda, com dashboard e IA como pontos fortes. A nota sobe para **8+** quando as anomalias críticas forem resolvidas e quando identidade visual, 404, termos no registro e tratamento de sessão estiverem alinhados a um produto “pronto para impressionar”.
