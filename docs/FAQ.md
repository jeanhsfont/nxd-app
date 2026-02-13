# ❓ FAQ - Perguntas Frequentes

## Conexão e Segurança

### Vocês vão entrar na minha rede interna?
**Não.** O módulo DX inicia a conexão de dentro para fora. Apenas dados saem da fábrica, ninguém entra.

### E se a internet da fábrica cair?
O DX armazena os dados localmente (buffer) e envia quando a conexão retornar.

### O dado que sai da fábrica é criptografado?
Sim, toda comunicação usa HTTPS (TLS 1.3).

### Preciso abrir alguma porta no firewall?
Não. O DX faz requisições HTTP de saída (porta 443), que normalmente já está liberada.

### A API Key pode ser roubada?
A chave tem 64 caracteres aleatórios. Você pode regenerá-la a qualquer momento no dashboard.

## Conectividade

### Quanto de dados o chip 4G consome por mês?
Aproximadamente 100-500 MB/mês por máquina, dependendo da frequência de envio.

### O DX funciona com CLPs da Siemens, Delta, Mitsubishi?
Sim! O DX é um gateway universal que traduz múltiplos protocolos.

### Quantas máquinas posso conectar em um único DX?
Até 32 máquinas simultaneamente, dependendo do modelo do DX.

### Posso usar Wi-Fi ou precisa ser cabo de rede?
Ambos funcionam. Para ambientes industriais, recomendamos cabo para maior estabilidade.

## Dados e Relatórios

### Eu consigo exportar esses dados para Excel?
Sim! (Funcionalidade em desenvolvimento)

### O sistema avisa se a temperatura passar de um limite?
Sim, você pode configurar alertas personalizados.

### Consigo ver o histórico de um ano atrás?
Sim, todos os dados são armazenados indefinidamente.

### O sistema identifica micro-paradas de 10 segundos?
Sim, a resolução temporal é de 1 segundo.

### Dá para integrar com o meu ERP atual?
Sim, oferecemos API REST completa para integração.

## Usabilidade

### Posso acessar pelo celular?
Sim, o dashboard é responsivo e funciona em qualquer dispositivo.

### Dá para mudar o nome das tags que o CLP manda?
Sim, você pode renomear tags no dashboard.

### Consigo criar alertas via WhatsApp ou E-mail?
Sim! (Funcionalidade em desenvolvimento)

### Quantos usuários podem estar logados ao mesmo tempo?
Ilimitado.

### Preciso de um servidor dentro da minha fábrica?
Não, tudo roda na nuvem.

## Custo e ROI

### Quanto eu vou economizar usando isso?
Clientes reportam redução de 15-30% em paradas não planejadas.

### Em quanto tempo o sistema se paga?
Média de 6-12 meses, dependendo do porte da operação.

### Se eu cancelar o serviço, os dados são meus?
Sim, você pode exportar todos os dados antes de cancelar.

### O sistema ajuda a prever quando a máquina vai quebrar?
Sim, com análise de tendências e machine learning (roadmap).

## Técnicas

### Onde os dados ficam hospedados?
Servidores na nuvem (AWS/Azure) com redundância.

### O sistema segue as regras da LGPD?
Sim, totalmente conforme.

### Qual a latência entre a máquina e a tela?
1-3 segundos em média (depende da conexão 4G).

### O sistema suporta protocolos antigos?
Sim, o DX traduz protocolos legados para formato moderno.

### Vocês oferecem API para PowerBI?
Sim, API REST completa com documentação.

## Suporte

### Como entro em contato com o suporte?
- Email: suporte@hubsystem.com.br
- WhatsApp: (XX) XXXX-XXXX
- Portal: https://suporte.hubsystem.com.br

### Vocês oferecem treinamento?
Sim, treinamento online e presencial disponível.

### Qual o SLA (tempo de resposta)?
- Crítico: 2 horas
- Alto: 8 horas
- Médio: 24 horas
- Baixo: 48 horas
