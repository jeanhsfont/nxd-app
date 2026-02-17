# NXD — Especificação para Google Cloud e Sistema Completo

## 1. Situação Atual (Aprovada)

### 1.1 Ambiente de teste

- **Melhor ambiente virtual possível:** CLPs simulados (Siemens + Delta) em Modbus TCP; DX Gateway (Node-RED/JS) lendo protocolo real e enviando HTTP para NXD.
- **Conclusão:** Sistema provado. Operação aprovada. Próximo passo é remodelar (fluxo e telas), depois beleza, e colocar no Google Cloud.

### 1.2 O que NÃO fazer agora

- **Login no NXD atual (Railway):** Não implementar. Será feito no ambiente Google Cloud.
- **Design/CSS elaborado:** Primeiro fluxo e funcionalidades; estilo depois.

---

## 2. Ordem de Trabalho

1. **Google Cloud** — Subir NXD no GCP (Cloud Run + Cloud SQL, etc.).
2. **Delimitação legal e técnica** — Contrato, termos de uso, boas práticas; o que é nossa responsabilidade e o que não é.
3. **Logs estruturados e integridade** — Garantir que o que chegou está registrado e não alterado; servidor seguro.
4. **API blindada** — Política de API Key (mostrar uma vez; se perder, gerar outra); responsabilidade do cliente ao copiar e colar no DX.
5. **Planejar e montar telas** — Fluxo completo: login, primeira tela (resumo), configurações, perfil, recuperação de senha, etc.
6. **Beleza** — Só depois do fluxo e das funcionalidades fechados.

---

## 3. Responsabilidades (Resumo para Contrato)

### 3.1 O NXD garante

- **Funcionamento do sistema:** disponibilidade (SLA), monitoramento e funcionalidades do NXD.
- **Integridade dos dados:** o que chega na nossa API é armazenado e exibido sem alteração.
- **Rastreabilidade:** logs estruturados do que foi recebido (origem, timestamp, payload).
- **Segurança do servidor:** dados em ambiente protegido (Google Cloud).

### 3.2 O NXD NÃO garante

- **Funcionamento das máquinas e sensores:** não configuramos CLP nem sensores; não garantimos calibração nem operação do chão de fábrica.
- **Precisão dos dados na origem:** garantimos que exibimos o que o DX envia; a precisão depende do CLP e da instalação do cliente.
- **Uso indevido da API Key:** após a geração, a chave é de responsabilidade do cliente (copiar uma vez, guardar no DX; se vazar ou perder, gerar outra).

### 3.3 Responsabilidade do cliente

- Manter equipamentos e sensores em condições adequadas.
- Copiar a API Key com segurança e configurá-la no DX; após isso, qualquer vazamento ou perda é de sua conta.
- Validar periodicamente a consistência dos dados na origem.
- Assinar termos de uso e política de contratação cientes dessas delimitações.

**Forma de garantir que os dados que chegam são do CLP:** logs estruturados (timestamp, origem, payload/hash). Não podemos “provar” fisicamente que veio do CLP, mas podemos provar o que **recebemos** e que não foi alterado em nosso lado.

---

## 4. API Key — Política “estilo Google”

- **Na criação:** a API Key é exibida **uma única vez** na tela.
- **Não há “copiar de novo”:** se o cliente perder, deve **gerar uma nova** (a antiga pode ser invalidada).
- **Efeito:** reduz vazamento por “segunda cópia”; deixa claro que guardar e levar ao DX é obrigação do cliente.
- **Contrato:** cliente declara que copiou a chave com segurança e a configurou no DX; problemas decorrentes de perda ou vazamento após isso não são atribuídos ao NXD.
- **Criptografia:** DX real usará HTTPS; autenticação em tempo real via API Key. Não precisamos “copiar” a criptografia do DX; a conexão TLS + API Key cobre o canal.

---

## 5. Logs e Integridade

- **Log estruturado por ingestão:** timestamp (UTC), IP de origem, API Key (hash ou ID, não texto plano), factory_id, device_id, tamanho do payload, hash do payload (opcional).
- **Objetivo:** provar o que chegou e que não foi alterado no nosso lado; suportar auditoria e eventual disputa.
- **Retenção:** definir política (ex.: 90 dias em log; dados de produção conforme contrato).
- **Servidor:** ambiente Google Cloud com boas práticas (rede, IAM, criptografia em repouso).

---

## 6. Modelo de Licença e Escopo

- **Uma fábrica por licença.** Uma conta = uma fábrica.
- **Outra fábrica:** nova licença, novo contrato NXD, outro login (outro navegador/app ou mesma conta com “troca de fábrica” no futuro).
- **Cenário 50 máquinas:** todas sob a **mesma fábrica**; um ou mais DX enviam para a mesma API Key dessa fábrica; o sistema recebe e organiza por device_id/máquina. Não precisa “50 licenças”.

---

## 7. Fluxo de Telas (Planejamento)

### 7.1 Antes do login

| Tela        | Conteúdo |
|------------|----------|
| Login      | Email + senha; “Esqueci minha senha”. |
| Recuperar senha | Email; link ou token para redefinir. |

### 7.2 Primeira tela após login (já configurado)

**Objetivo:** resumo global para o gestor (ex.: em férias, abre o app e vê em 5 segundos se está tudo ok).

- **Máquinas online** (ex.: 48/50).
- **Nível de produção** vs média histórica (ex.: “Dentro da média” ou “X% acima/abaixo”).
- **Consumo de energia** vs média (ex.: “Dentro da média”).
- **Indicadores visuais simples:** verde = dentro da média/normal; amarelo/vermelho = atenção.
- **Mensagem de contexto:** “Tudo operando dentro da média” ou “Verifique as máquinas offline”.
- **Atalhos:** para Live View, Financeiro, Comparativos, Configurações.

Não é tela de “criar fábrica”; quem já tem fábrica cai direto nesse resumo.

### 7.3 Telas principais (após o resumo)

| Tela          | Conteúdo |
|---------------|----------|
| Live View     | Monitoramento em tempo real das máquinas (já existente; adaptar). |
| Financeiro    | Lucro cessante, produção, custos (já existente; adaptar). |
| Comparativos  | Comparar máquinas (ex.: A vs B; filtros por período). Configuração: “como comparar” (quais métricas, quais máquinas). |
| Histórico / Relatórios | Gráficos e tabelas por período; export (futuro). |

### 7.4 Configurações (organizadas)

- **Perfil do cliente**  
  Nome, email, telefone, endereço da fábrica (o que fizer sentido).

- **Configurações da fábrica**  
  Nome da fábrica, fuso horário, etc.

- **API Key**  
  Ver política acima: gerar nova (mostrar uma vez); aviso de responsabilidade ao copiar e levar ao DX.

- **Configurações técnicas / sistema**  
  - Velocidade de atualização (polling/refresh).  
  - Buffer/velocidade de envio (se aplicável ao que o cliente pode configurar).  
  - Outros parâmetros que forem expostos (ex.: timezone para relatórios).

- **Configurações de comparação**  
  Quais máquinas comparar; quais métricas (produção, energia, etc.).

- **Configurações de alertas**  
  (Futuro: limites, canais, etc.)

Tudo que hoje está espalhado em “configurações” deve ficar agrupado de forma lógica (Perfil, Fábrica, API, Sistema, Comparação, Alertas).

### 7.5 Telas padrão de conta

- Termos de uso / Política de uso (link no cadastro e no rodapé).
- Política de privacidade (LGPD).
- Contrato / condições de contratação (ao assinar o serviço).

---

## 8. Cenário: 50 máquinas, CLP + DX

- **Recepção:** um ou mais DX enviam para a **mesma API Key** (uma fábrica). Cada mensagem traz `device_id` (e opcionalmente brand, protocol).  
- **Auto-discovery:** NXD cria/atualiza máquinas por `device_id`; até 50 (ou mais) máquinas na mesma fábrica.  
- **Primeira tela:** mostra totais (ex.: 48 online, 2 offline) e médias globais (produção, energia).  
- **Demais telas:** filtros por máquina, grupos ou “todas”; comparação máquina a máquina configurável.

Não há “50 contratos”; é 1 fábrica com 50 máquinas.

---

## 9. Legal — O que pode e o que não pode (Parâmetro geral)

- **Pode:** limitar responsabilidade ao que o sistema controla (disponibilidade, integridade, rastreabilidade); exigir que o cliente mantenha equipamentos e cuide da API Key; exigir aceite de termos e política antes do uso.  
- **Não pode:** abusar (ex.: isenção total sem contrapartida); esconder limitações; deixar de oferecer mínimo de segurança e documentação.  
- **Recomendação:** contrato e termos redigidos/revistos por advogado; este doc serve como base técnica e de escopo para o jurídico.

---

## 10. Checklist de Entrega (Ordem)

1. [ ] Especificação de telas e fluxos (este doc) aprovada.  
2. [ ] Deploy NXD no Google Cloud (Cloud Run, Cloud SQL, logs).  
3. [ ] Logs estruturados e política de retenção.  
4. [ ] Política de API Key (mostrar uma vez; gerar nova se perder).  
5. [ ] Rascunho de termos de uso e contrato (para advogado).  
6. [ ] Implementar fluxo: Login → Primeira tela (resumo) → Live View, Financeiro, Comparativos → Configurações (perfil, fábrica, API, sistema, comparação).  
7. [ ] Recuperação de senha.  
8. [ ] Revisão de layout/organização (sem estilo final ainda).  
9. [ ] Design/beleza depois que o fluxo estiver redondo.

---

*Documento de referência para desenvolvimento e jurídico. Atualizar conforme decisões de produto e jurídico.*
