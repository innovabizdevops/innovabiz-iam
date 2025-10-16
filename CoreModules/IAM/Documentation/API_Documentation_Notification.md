# API de Notificações - Documentação

## Introdução

A API de Notificações do módulo IAM da plataforma InnovaBiz fornece interfaces para envio, gerenciamento e rastreamento de notificações através de múltiplos canais. Esta documentação descreve as interfaces, métodos, parâmetros e respostas disponíveis.

## Interfaces Principais

### NotificationService

Interface principal para envio de notificações através de múltiplos canais.

#### Métodos

##### `send(recipient, content, options?, event?): Promise<SendResult>`

Envia uma notificação para um destinatário específico.

**Parâmetros:**
- `recipient: NotificationRecipient` - Destinatário da notificação
- `content: string | RenderedTemplate` - Conteúdo da notificação ou template renderizado
- `options?: SendOptions` - Opções para envio (opcional)
- `event?: BaseEvent` - Evento que originou a notificação (opcional)

**Retorno:** `Promise<SendResult>` - Resultado do envio

**Exemplo:**
```typescript
const result = await notificationService.send(
  { 
    id: 'user123', 
    metadata: { 
      email: 'usuario@exemplo.com',
      name: 'João Silva' 
    } 
  },
  'Olá João, sua conta foi criada com sucesso!',
  {
    preferredChannels: [NotificationChannel.EMAIL]
  }
);
```

##### `sendBatch(recipients, content, options?, event?): Promise<BatchSendResult>`

Envia a mesma notificação para múltiplos destinatários.

**Parâmetros:**
- `recipients: NotificationRecipient[]` - Lista de destinatários
- `content: string | RenderedTemplate` - Conteúdo da notificação
- `options?: SendOptions` - Opções para envio (opcional)
- `event?: BaseEvent` - Evento que originou a notificação (opcional)

**Retorno:** `Promise<BatchSendResult>` - Resultados do envio em lote

**Exemplo:**
```typescript
const results = await notificationService.sendBatch(
  [
    { id: 'user1', metadata: { email: 'user1@example.com' } },
    { id: 'user2', metadata: { email: 'user2@example.com' } }
  ],
  'Novo recurso disponível: conheça nossa plataforma renovada!',
  {
    preferredChannels: [NotificationChannel.EMAIL]
  }
);
```

##### `scheduleNotification(recipient, content, scheduledTime, options?): Promise<ScheduleResult>`

Agenda uma notificação para envio futuro.

**Parâmetros:**
- `recipient: NotificationRecipient` - Destinatário da notificação
- `content: string | RenderedTemplate` - Conteúdo da notificação
- `scheduledTime: Date` - Momento para envio
- `options?: SendOptions` - Opções para envio (opcional)

**Retorno:** `Promise<ScheduleResult>` - Resultado do agendamento

**Exemplo:**
```typescript
const tomorrow = new Date();
tomorrow.setDate(tomorrow.getDate() + 1);

const result = await notificationService.scheduleNotification(
  { id: 'user123', metadata: { email: 'usuario@exemplo.com' } },
  'Lembrete: seu evento começa amanhã!',
  tomorrow,
  { preferredChannels: [NotificationChannel.EMAIL, NotificationChannel.SMS] }
);
```

##### `cancelNotification(notificationId): Promise<boolean>`

Cancela uma notificação agendada que ainda não foi enviada.

**Parâmetros:**
- `notificationId: string` - ID da notificação

**Retorno:** `Promise<boolean>` - Indica sucesso do cancelamento

**Exemplo:**
```typescript
const cancelled = await notificationService.cancelNotification('notif-123456');
```

##### `getNotificationStatus(notificationId): Promise<NotificationStatusInfo>`

Obtém o status atual de uma notificação.

**Parâmetros:**
- `notificationId: string` - ID da notificação

**Retorno:** `Promise<NotificationStatusInfo>` - Informações de status

**Exemplo:**
```typescript
const status = await notificationService.getNotificationStatus('notif-123456');
```

### TemplateService

Interface para gerenciamento e renderização de templates de notificação.

#### Métodos

##### `render(templateId, options): Promise<RenderedTemplate>`

Renderiza um template com variáveis específicas.

**Parâmetros:**
- `templateId: string` - ID do template
- `options: RenderOptions` - Opções de renderização

**Retorno:** `Promise<RenderedTemplate>` - Template renderizado

**Exemplo:**
```typescript
const rendered = await templateService.render('welcome-email', {
  targetChannel: NotificationChannel.EMAIL,
  locale: 'pt-BR',
  variables: {
    name: 'João Silva',
    activationLink: 'https://exemplo.com/activate/abc123'
  }
});
```

##### `createTemplate(template): Promise<string>`

Cria um novo template.

**Parâmetros:**
- `template: NotificationTemplate` - Definição do template

**Retorno:** `Promise<string>` - ID do template criado

**Exemplo:**
```typescript
const templateId = await templateService.createTemplate({
  name: 'welcome-email',
  category: 'onboarding',
  content: {
    'pt-BR': {
      subject: 'Bem-vindo à InnovaBiz',
      body: 'Olá {{name}}, seja bem-vindo à InnovaBiz!'
    },
    'en-US': {
      subject: 'Welcome to InnovaBiz',
      body: 'Hello {{name}}, welcome to InnovaBiz!'
    }
  }
});
```

##### `updateTemplate(templateId, template): Promise<boolean>`

Atualiza um template existente.

**Parâmetros:**
- `templateId: string` - ID do template
- `template: Partial<NotificationTemplate>` - Campos a atualizar

**Retorno:** `Promise<boolean>` - Indica sucesso da atualização

**Exemplo:**
```typescript
const updated = await templateService.updateTemplate('welcome-email', {
  content: {
    'pt-BR': {
      subject: 'Bem-vindo à nova InnovaBiz',
      body: 'Olá {{name}}, seja bem-vindo à nova plataforma InnovaBiz!'
    }
  }
});
```

##### `getTemplate(templateId): Promise<NotificationTemplate>`

Obtém um template pelo ID.

**Parâmetros:**
- `templateId: string` - ID do template

**Retorno:** `Promise<NotificationTemplate>` - Template encontrado

**Exemplo:**
```typescript
const template = await templateService.getTemplate('welcome-email');
```

##### `listTemplates(filter?): Promise<NotificationTemplateSummary[]>`

Lista templates disponíveis com filtros opcionais.

**Parâmetros:**
- `filter?: TemplateFilter` - Filtros (opcional)

**Retorno:** `Promise<NotificationTemplateSummary[]>` - Lista de templates

**Exemplo:**
```typescript
const templates = await templateService.listTemplates({
  category: 'onboarding',
  active: true
});
```

### NotificationTrackingService

Interface para rastreamento e análise de notificações.

#### Métodos

##### `trackSend(notificationId, recipientId, channel, provider?, metadata?): Promise<void>`

Registra evento de envio de notificação.

**Parâmetros:**
- `notificationId: string` - ID da notificação
- `recipientId: string` - ID do destinatário
- `channel: NotificationChannel` - Canal utilizado
- `provider?: string` - Provedor de serviço (opcional)
- `metadata?: Record<string, any>` - Metadados adicionais (opcional)

**Exemplo:**
```typescript
await trackingService.trackSend(
  'notif-123456',
  'user123',
  NotificationChannel.EMAIL,
  'sendgrid'
);
```

##### `trackDelivery(notificationId, recipientId, channel, deliveryData?): Promise<void>`

Registra evento de entrega de notificação.

**Parâmetros:**
- `notificationId: string` - ID da notificação
- `recipientId: string` - ID do destinatário
- `channel: NotificationChannel` - Canal utilizado
- `deliveryData?: object` - Dados de entrega (opcional)

**Exemplo:**
```typescript
await trackingService.trackDelivery(
  'notif-123456',
  'user123',
  NotificationChannel.EMAIL,
  {
    timestamp: new Date(),
    responseId: 'resp-789'
  }
);
```

##### `trackOpen(notificationId, recipientId, channel, openData?): Promise<void>`

Registra evento de abertura de notificação.

**Parâmetros:**
- `notificationId: string` - ID da notificação
- `recipientId: string` - ID do destinatário
- `channel: NotificationChannel` - Canal utilizado
- `openData?: object` - Dados de abertura (opcional)

**Exemplo:**
```typescript
await trackingService.trackOpen(
  'notif-123456',
  'user123',
  NotificationChannel.EMAIL,
  {
    userAgent: 'Mozilla/5.0...',
    ipAddress: '192.168.1.1'
  }
);
```

##### `trackClick(notificationId, recipientId, channel, clickData): Promise<void>`

Registra evento de clique em link de notificação.

**Parâmetros:**
- `notificationId: string` - ID da notificação
- `recipientId: string` - ID do destinatário
- `channel: NotificationChannel` - Canal utilizado
- `clickData: object` - Dados do clique

**Exemplo:**
```typescript
await trackingService.trackClick(
  'notif-123456',
  'user123',
  NotificationChannel.EMAIL,
  {
    url: 'https://exemplo.com/activate',
    userAgent: 'Mozilla/5.0...',
    ipAddress: '192.168.1.1'
  }
);
```

##### `getTrackingEvents(notificationId): Promise<NotificationTrackingEvent[]>`

Obtém histórico de eventos para uma notificação.

**Parâmetros:**
- `notificationId: string` - ID da notificação

**Retorno:** `Promise<NotificationTrackingEvent[]>` - Lista de eventos

**Exemplo:**
```typescript
const events = await trackingService.getTrackingEvents('notif-123456');
```

##### `getTrackingSummary(notificationId): Promise<NotificationTrackingSummary>`

Obtém resumo de rastreamento para uma notificação.

**Parâmetros:**
- `notificationId: string` - ID da notificação

**Retorno:** `Promise<NotificationTrackingSummary>` - Resumo de rastreamento

**Exemplo:**
```typescript
const summary = await trackingService.getTrackingSummary('notif-123456');
```

##### `getAggregateStats(filter): Promise<NotificationAggregateStats>`

Obtém estatísticas agregadas com filtros.

**Parâmetros:**
- `filter: StatsFilter` - Filtros para as estatísticas

**Retorno:** `Promise<NotificationAggregateStats>` - Estatísticas agregadas

**Exemplo:**
```typescript
const stats = await trackingService.getAggregateStats({
  startDate: new Date('2023-01-01'),
  endDate: new Date('2023-01-31'),
  channel: NotificationChannel.EMAIL,
  module: 'iam'
});
```

### NotificationIntegrationService

Interface para integração do sistema de notificações com outros módulos.

#### Métodos

##### `processEvent(event): Promise<any>`

Processa um evento do sistema para notificação.

**Parâmetros:**
- `event: BaseEvent` - Evento a ser processado

**Retorno:** `Promise<any>` - Resultado do processamento

**Exemplo:**
```typescript
await integrationService.processEvent({
  id: 'evt-123456',
  timestamp: new Date(),
  module: 'iam',
  category: 'authentication',
  type: 'login-success',
  data: {
    userId: 'user123',
    userName: 'João Silva',
    userEmail: 'joao@example.com'
  }
});
```

##### `registerEventProcessor(module, category, processor): void`

Registra um processador para eventos específicos.

**Parâmetros:**
- `module: string` - Módulo de origem
- `category: string` - Categoria do evento
- `processor: EventProcessor` - Função processadora

**Exemplo:**
```typescript
integrationService.registerEventProcessor(
  'payment-gateway',
  'refund',
  async (event) => {
    // Lógica específica para processar eventos de reembolso
    // ...
    return result;
  }
);
```

## Tipos de Dados

### NotificationChannel (enum)

```typescript
enum NotificationChannel {
  EMAIL = 'email',
  SMS = 'sms',
  PUSH = 'push',
  WEBHOOK = 'webhook'
}
```

### NotificationRecipient (interface)

```typescript
interface NotificationRecipient {
  id: string;
  metadata?: {
    email?: string;
    phone?: string;
    deviceToken?: string;
    webhookUrl?: string;
    name?: string;
    preferredChannels?: NotificationChannel[];
    preferredLanguage?: string;
    [key: string]: any;
  };
}
```

### SendOptions (interface)

```typescript
interface SendOptions {
  notificationId?: string;
  preferredChannels?: NotificationChannel[];
  simultaneousChannels?: boolean;
  priority?: 'high' | 'medium' | 'low';
  retryStrategy?: {
    maxAttempts: number;
    initialDelayMs: number;
    maxDelayMs: number;
  };
  tracking?: {
    source?: string;
    category?: string;
    tags?: string[];
    metadata?: Record<string, any>;
  };
  variables?: Record<string, any>;
}
```

### SendResult (interface)

```typescript
interface SendResult {
  success: boolean;
  notificationId: string;
  channel?: NotificationChannel;
  provider?: string;
  timestamp: Date;
  recipientId: string;
  statusCode?: string;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  providerResponse?: any;
  trackingInfo?: {
    trackingId: string;
    trackingUrl?: string;
  };
}
```

### NotificationStatus (enum)

```typescript
enum NotificationStatus {
  SCHEDULED = 'SCHEDULED',
  SENDING = 'SENDING',
  SENT = 'SENT',
  DELIVERED = 'DELIVERED',
  FAILED = 'FAILED',
  OPENED = 'OPENED',
  CLICKED = 'CLICKED',
  REPLIED = 'REPLIED',
  BOUNCED = 'BOUNCED',
  REJECTED = 'REJECTED',
  SPAM = 'SPAM',
  EXPIRED = 'EXPIRED',
  CANCELLED = 'CANCELLED',
  UNKNOWN = 'UNKNOWN'
}
```

### BaseEvent (interface)

```typescript
interface BaseEvent {
  id: string;
  timestamp: Date;
  module: string;
  category: string;
  type?: string;
  data?: Record<string, any>;
  context?: {
    tenantId?: string;
    environment?: string;
    source?: string;
    locale?: string;
    metadata?: Record<string, any>;
  };
  severity?: 'low' | 'medium' | 'high' | 'critical';
  recipients?: NotificationRecipient[];
  metadata?: Record<string, any>;
}
```

## Códigos de Erro

| Código | Descrição | Ação Recomendada |
|--------|-----------|------------------|
| `INVALID_RECIPIENT` | Dados do destinatário inválidos ou insuficientes | Verificar se os dados do destinatário estão completos |
| `INVALID_CONTENT` | Conteúdo da notificação inválido | Verificar o conteúdo da notificação |
| `TEMPLATE_NOT_FOUND` | Template não encontrado | Verificar ID do template |
| `CHANNEL_UNAVAILABLE` | Canal de notificação indisponível | Tentar canal alternativo ou verificar status do provedor |
| `PROVIDER_ERROR` | Erro do provedor de serviço | Verificar logs para detalhes específicos |
| `SENDING_LIMIT_EXCEEDED` | Limite de envios excedido | Aguardar e tentar novamente mais tarde |
| `RECIPIENT_UNSUBSCRIBED` | Destinatário cancelou assinatura | Respeitar preferência do usuário |
| `INVALID_CONFIGURATION` | Configuração inválida | Verificar configuração do sistema |
| `TRACKING_ERROR` | Erro ao rastrear notificação | Verificar configuração de rastreamento |
| `SCHEDULING_ERROR` | Erro ao agendar notificação | Verificar dados de agendamento |

## Webhooks

O sistema de notificações pode receber webhooks de provedores externos para atualizar o status das notificações. Os endpoints para webhooks são:

### `/api/notifications/webhooks/{provider}`

**Método:** POST

**Parâmetros de URL:**
- `provider` - Nome do provedor (ex: sendgrid, twilio)

**Corpo da Requisição:**
Varia conforme o provedor.

**Resposta de Sucesso:**
```json
{
  "success": true,
  "processed": 1
}
```

**Resposta de Erro:**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_WEBHOOK",
    "message": "Invalid webhook payload"
  }
}
```

## Endpoints Públicos de Rastreamento

### `/t/{notificationId}/{recipientId}/{channel}/{linkId}`

Endpoint para rastreamento de cliques em links.

**Método:** GET

**Parâmetros de URL:**
- `notificationId` - ID da notificação
- `recipientId` - ID do destinatário
- `channel` - Canal utilizado
- `linkId` - ID do link

**Resposta:** Redirecionamento para URL original

### `/p/{notificationId}/{recipientId}/{channel}/{trackingId}.gif`

Endpoint para rastreamento de abertura via pixel transparente.

**Método:** GET

**Parâmetros de URL:**
- `notificationId` - ID da notificação
- `recipientId` - ID do destinatário
- `channel` - Canal utilizado
- `trackingId` - ID de rastreamento

**Resposta:** Imagem GIF transparente 1x1