# Plano de execução — 12 semanas (NXD vendável)

Objetivo: sair de “promissor” para “comprável por indústria média/grande” em 90 dias.

---

## Visão por prioridade

| # | Prioridade | Objetivo em uma frase |
|---|------------|------------------------|
| 1 | ROI / valor financeiro | Dashboard executivo com ROI real, sem mock, exportação PDF/CSV |
| 2 | Segurança e compliance | 2FA no login, RBAC, audit log, JWT/secrets seguros |
| 3 | Billing real | Checkout, webhooks, assinatura, limites por plano |
| 4 | IA produto | Relatórios auditáveis, fontes citadas, histórico |
| 5 | UX enterprise | Linguagem/copy, design system, onboarding “executivo” |

---

## Sprint 1–2 (Semanas 1–4): ROI e dados reais

### 1.1 Backend: cálculos financeiros reais

**Backlog**
- Trocar qualquer retorno mock de financeiro por cálculo a partir de telemetria + `business_config` / `tag_mappings`.
- Endpoints que retornem, por período (hoje, 24h, 7d, 30d) e por setor/ativo:
  - faturamento (peças OK × valor_venda_ok),
  - perda refugo (peças NOK × custo_refugo_un),
  - custo de parada (horas_parada × custo_parada_h),
  - “perdas evitadas” / “custo de parada evitado” (comparativo período anterior ou meta).
- Tendência semanal/mensal (agregados por semana/mês).
- Garantir que tudo use `factory_id` / `sector_id` e regras por fábrica/setor.

**Critérios de aceite**
- Nenhum dado financeiro no dashboard vem de mock.
- Valores batem com telemetria ingerida e config de negócio da fábrica/setor.
- Resposta de API documentada (ex.: OpenAPI ou README de endpoints).

**Definition of Done**
- Código em main, testes automatizados para os cálculos principais, revisão feita.

---

### 1.2 Frontend: dashboard executivo + exportação

**Backlog**
- Cards/seções no dashboard: perdas evitadas, custo de parada evitado, ganho por setor/ativo, tendência semanal/mensal.
- Filtros: período (hoje, 7d, 30d), setor (e “todos”).
- Exportação:
  - CSV: tabela de resumo (período, setor, métricas) para a diretoria.
  - PDF: uma página “resumo executivo” com os mesmos indicadores e período selecionado.

**Critérios de aceite**
- Dashboard mostra apenas dados vindos da API (sem mock).
- Usuário consegue exportar CSV e PDF com os dados exibidos no período/setor selecionado.
- Textos e rótulos alinhados a “painel de decisão financeira” (ex.: “ROI”, “Perdas evitadas”, “Custo de parada”).

**Definition of Done**
- Feature em main, testada em staging, copy revisado.

---

## Sprint 3–4 (Semanas 5–8): Segurança e compliance

### 2.1 2FA no fluxo de login

**Backlog**
- Após login com senha, se o usuário tiver 2FA ativo: redirecionar para tela de código TOTP (ou link “Enviar código por e-mail” se houver).
- Backend: retornar `requires_2fa: true` + token temporário ou sessão intermediária; após validar TOTP, emitir JWT final.
- Frontend: tela de “Digite o código do app” (e opção “Configurar 2FA” em Configurações já existente).

**Critérios de aceite**
- Login com 2FA ativado exige código TOTP antes de acessar o app.
- Sem 2FA, fluxo atual de login permanece.
- Documentação mínima (para o time) de como ativar e testar 2FA.

**Definition of Done**
- Fluxo E2E testado (ativar 2FA → logout → login com código), código em main.

---

### 2.2 JWT e secrets em produção

**Backlog**
- Remover fallback de JWT_SECRET em código (ex.: “your-secret-key-change-in-production”).
- Exigir `JWT_SECRET` (e demais secrets) via variável de ambiente / Secret Manager; falhar no boot se ausente em produção.
- Documentar no README ou em runbook: “Produção exige JWT_SECRET (e X, Y) configurados”.

**Critérios de aceite**
- Em ambiente de produção, app não sobe sem JWT_SECRET (e secrets definidos) configurados.
- Nenhum valor default inseguro em código para produção.

**Definition of Done**
- Deploy de produção validado com secrets externos; doc atualizada.

---

### 2.3 Trilha de auditoria (audit log)

**Backlog**
- Tabela (ou schema) de audit_log: `user_id`, `action`, `entity_type`, `entity_id`, `old_value` (JSON opcional), `new_value` (JSON opcional), `ip`, `created_at`.
- Registrar: alteração de setor, alteração de ativo, alteração de configs de negócio, alteração de usuário/perfil (quando houver).
- Endpoint (admin): listar audit log com filtros (usuário, entidade, período).

**Critérios de aceite**
- Todas as alterações listadas acima geram um registro de auditoria.
- Admin consegue consultar quem alterou o quê e quando (via API ou tela interna).

**Definition of Done**
- Migração aplicada, eventos principais instrumentados, endpoint/tela de consulta em main.

---

### 2.4 RBAC (perfis: admin, operador, gestor, financeiro)

**Backlog**
- Modelo: `role` por usuário (ou por usuário+fábrica): admin, operador, gestor, financeiro.
- Regras: quem pode ver/editar setores, ativos, config financeira, billing, usuários, audit log.
- Backend: middleware ou checagem de role em cada rota sensível.
- Frontend: esconder ou desabilitar ações conforme o role do usuário logado.

**Critérios de aceite**
- Pelo menos 4 perfis com regras claras (documentadas).
- Acesso negado (403) quando role não permitir a ação.
- UI não mostra opções que o usuário não pode executar.

**Definition of Done**
- Matriz de permissões documentada, testes de acesso, código em main.

---

## Sprint 5–6 (Semanas 9–10): Billing real

### 3.1 Gateway e checkout

**Backlog**
- Escolher um gateway (Stripe, Pagar.me ou Iugu) e criar conta/ambiente de teste.
- Backend: criar assinatura, plano (trial, básico, pro, etc.), preços; endpoint de “checkout session” ou link de pagamento.
- Frontend: tela de planos (já existe em parte) com CTA “Assinar” levando ao checkout real (redirect ao gateway).
- Webhook do gateway: receber evento de pagamento aprovado/falha e atualizar status da assinatura no banco.

**Critérios de aceite**
- Usuário consegue escolher plano e completar um pagamento de teste no gateway.
- Status da assinatura (ativa, trial, inadimplente) é persistido e atualizado via webhook.
- Não há “mock” de plano em produção (dados vêm do gateway/banco).

**Definition of Done**
- Fluxo de assinatura E2E em ambiente de teste, webhook tratado, doc de configuração do gateway.

---

### 3.2 Lifecycle: trial, upgrade/downgrade, inadimplência

**Backlog**
- Trial: N dias grátis; ao fim, exibir aviso e bloquear acesso se não houver assinatura paga.
- Upgrade/downgrade: troca de plano com proration (conforme API do gateway); atualizar limites no sistema.
- Inadimplência: ao receber webhook de falha, marcar assinatura como “past_due” ou “inadimplente”; notificar usuário (e-mail) e aplicar bloqueio parcial (ex.: só leitura) ou total conforme política.
- Nota/recibo: usar emissão do gateway ou integrar com serviço de nota; guardar link ou PDF no perfil do usuário.

**Critérios de aceite**
- Trial expira e bloqueia (ou avisa) conforme regra definida.
- Upgrade/downgrade refletido no produto e na fatura.
- Inadimplência resulta em estado visível e restrição de uso conforme definido.
- Usuário tem acesso a comprovante/nota (link ou PDF).

**Definition of Done**
- Cenários principais testados, documentação de políticas (trial, bloqueio) e código em main.

---

## Sprint 7 (Semanas 11): IA produto e UX

### 4.1 IA com relatórios auditáveis

**Backlog**
- Todas as respostas da IA citam fontes (“baseado em X”: setor, período, métricas usadas).
- Histórico de análises: salvar relatório (texto + fontes + data + responsável) no banco; listar na UI.
- Templates: pelo menos um “Relatório executivo semanal” (estrutura fixa, dados reais preenchidos).
- Versionamento: cada relatório gerado tem id e data; possível “ver versão anterior”.

**Critérios de aceite**
- Relatório exibido na UI mostra “Baseado em: …” (dados/periodo usados).
- Usuário vê histórico de relatórios e pode abrir um anterior.
- Um template de relatório executivo está disponível e preenchido com dados reais.

**Definition of Done**
- Persistência de relatórios, tela de histórico, template executivo e código em main.

---

### 5.1 Refino UX enterprise

**Backlog**
- Padronizar textos: toasts, empty states, mensagens de erro e loading (linguagem enterprise, sem jargão técnico).
- Design system: garantir tokens (cores, espaçamento) e componentes (botão, input, card) consistentes em todas as telas.
- Onboarding: reduzir passos técnicos; focar em “resultado” (ex.: “Você verá seu primeiro ROI em X minutos”).
- HTTP: usar apenas o cliente `api` (axios) para chamadas à API; remover `fetch` solto onde ainda existir.

**Critérios de aceite**
- Copy revisado em telas principais (login, onboarding, dashboard, erro).
- Nenhuma tela quebra o design system (cores, botões, inputs).
- Onboarding não exige conhecimento técnico para concluir.
- Todas as chamadas à API passam pelo mesmo cliente (ex.: `api.get/post`).

**Definition of Done**
- Checklist de telas aprovada, fetch substituído por `api`, código em main.

---

## Sprint 8 (Semana 12): Observabilidade e fechamento

### 6.1 Health e observabilidade

**Backlog**
- Health endpoint público: `GET /api/health` com status do app e, se possível, do banco (sem dados sensíveis).
- Status page (página estática ou serviço externo) com “Operacional” / “Degradado” e histórico de incidentes (manual no início).
- Logs estruturados (JSON) com request_id e nível (info, error); opcional: métricas (latência, erros) para SLA.

**Critérios de aceite**
- `/api/health` responde 200 quando o serviço está ok.
- Existe uma URL de “status” (interna ou externa) para a diretoria/operacao.
- Logs permitem rastrear um request e erros em produção.

**Definition of Done**
- Health e status documentados, exemplo de log em produção verificado.

---

### 6.2 Backlog “próxima onda” (pós-12 semanas)

Documentar como próximos passos (não obrigatórios no primeiro ciclo):

- Multi-tenant robusto (empresa > fábricas > setores > usuários).
- Alertas (WhatsApp/email/Slack) com escalonamento.
- Relatórios executivos agendados (semanal/mensal).
- Backups e política de retenção.
- Exportação e integrações ERP/MES/BI.

---

## Resumo por sprint

| Sprint | Semanas | Foco principal |
|--------|---------|----------------|
| 1–2 | 1–4 | ROI real + dashboard executivo + exportação PDF/CSV |
| 3–4 | 5–8 | 2FA no login, JWT/secrets, audit log, RBAC |
| 5–6 | 9–10 | Billing: checkout, webhooks, trial, inadimplência, nota |
| 7 | 11 | IA auditável + UX enterprise (copy, design system, onboarding, api única) |
| 8 | 12 | Health, status, logs; fechamento e doc “próxima onda” |

---

## Pedidos objetivos para o dev (recorte por bloco)

1. **ROI:** “Dashboard executivo com ROI real por período e por setor, sem dados mock, e com exportação PDF/CSV para diretoria.”
2. **Segurança:** “Hardening para produção: 2FA no fluxo de autenticação, RBAC, audit log e políticas de senha/sessão (JWT/secrets por ambiente).”
3. **Billing:** “Billing ponta a ponta com checkout, webhooks, status de assinatura e limites de uso por plano.”
4. **IA:** “IA com relatórios auditáveis, fontes citadas e histórico — pronto para reunião com diretoria.”
5. **UX:** “Refino de UX com linguagem enterprise, consistência visual e fluxo sem atrito do onboarding ao primeiro valor.”

---

## Definition of Done (geral)

- Código em `main` (ou branch de release acordada).
- Testes automatizados para regras críticas (cálculo financeiro, auth, billing).
- Sem mocks em produção para dados que devem ser reais (financeiro, plano).
- Documentação mínima (README, runbook ou doc de API) atualizada quando aplicável.
- Revisão de código feita.

Se quiser, no próximo passo dá para detalhar um desses blocos em tarefas técnicas (issues) ou em critérios de aceite ainda mais granulares por tela/endpoint.
