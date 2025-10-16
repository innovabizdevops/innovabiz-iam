/**
 * @file notification-template.ts
 * @description Sistema de templates para notificações
 * 
 * Este sistema permite a criação, gerenciamento e renderização de templates
 * de notificação com suporte a personalização e internacionalização.
 */

import * as fs from 'fs';
import * as path from 'path';
import * as handlebars from 'handlebars';
import { Logger } from '../../../infrastructure/observability/logger';
import { NotificationChannel } from './notification-channel';
import { NotificationContent } from '../adapters/notification-adapter';

/**
 * Linguagem ou localidade para internacionalização
 */
export type Locale = string;

/**
 * Interface para representar um modelo de template de notificação
 */
export interface TemplateModel {
  /**
   * Identificador único do template
   */
  id: string;
  
  /**
   * Versão do template
   */
  version: string;
  
  /**
   * Nome descritivo do template
   */
  name: string;
  
  /**
   * Descrição do template
   */
  description?: string;
  
  /**
   * Categorias para organização e busca
   */
  categories?: string[];
  
  /**
   * Templates específicos por canal
   */
  channelTemplates: Map<NotificationChannel, ChannelTemplate>;
  
  /**
   * Dados específicos por localidade
   */
  localizedData: Map<Locale, LocalizedTemplateData>;
  
  /**
   * Dados padrão para fallback
   */
  defaultData: LocalizedTemplateData;
  
  /**
   * Metadados adicionais
   */
  metadata?: Record<string, any>;
  
  /**
   * Data de criação
   */
  createdAt: Date;
  
  /**
   * Última modificação
   */
  updatedAt: Date;
  
  /**
   * Indica se o template está ativo
   */
  active: boolean;
}

/**
 * Interface para template específico por canal
 */
export interface ChannelTemplate {
  /**
   * Conteúdo template para título/assunto
   */
  subject?: string;
  
  /**
   * Conteúdo template para corpo principal (texto)
   */
  body: string;
  
  /**
   * Conteúdo template para corpo HTML (quando suportado)
   */
  htmlBody?: string;
  
  /**
   * Templates para componentes específicos do canal
   * Ex: botões, imagens, anexos, etc.
   */
  components?: Record<string, string>;
  
  /**
   * Configurações específicas para o canal
   */
  settings?: Record<string, any>;
}

/**
 * Interface para dados localizados de template
 */
export interface LocalizedTemplateData {
  /**
   * Linguagem/localidade
   */
  locale: Locale;
  
  /**
   * Textos estáticos específicos da localidade
   */
  strings: Record<string, string>;
  
  /**
   * Formatos específicos da localidade
   * (datas, moedas, números, etc)
   */
  formats: {
    date?: string;
    time?: string;
    currency?: string;
    number?: string;
    [key: string]: string | undefined;
  };
  
  /**
   * Preferências de mídia específicas da localidade
   */
  mediaPreferences?: {
    images?: string[];
    colors?: string[];
    icons?: string[];
    [key: string]: string[] | undefined;
  };
  
  /**
   * Configurações específicas da localidade
   */
  settings?: Record<string, any>;
}

/**
 * Opções para renderização de template
 */
export interface TemplateRenderOptions {
  /**
   * Localidade para renderização
   */
  locale?: Locale;
  
  /**
   * Canal de notificação alvo
   */
  targetChannel: NotificationChannel;
  
  /**
   * Variáveis para interpolação
   */
  variables: Record<string, any>;
  
  /**
   * Filtros personalizados
   */
  filters?: Record<string, Function>;
  
  /**
   * Opções específicas para transformação
   */
  transformOptions?: {
    /**
     * Substituir URLs para rastreamento
     */
    trackUrls?: boolean;
    
    /**
     * Adicionar pixel de rastreamento em HTML
     */
    addTrackingPixel?: boolean;
    
    /**
     * Converter markdown para HTML
     */
    markdownToHtml?: boolean;
    
    /**
     * Truncar conteúdo que excede limites do canal
     */
    truncateToChannelLimits?: boolean;
  };
}

/**
 * Resultados da renderização de template
 */
export interface RenderedTemplate {
  /**
   * Conteúdo renderizado
   */
  content: NotificationContent;
  
  /**
   * Template utilizado
   */
  templateId: string;
  
  /**
   * Versão do template
   */
  templateVersion: string;
  
  /**
   * Canal utilizado
   */
  channel: NotificationChannel;
  
  /**
   * Localidade utilizada
   */
  locale: Locale;
  
  /**
   * Variáveis utilizadas
   */
  variables: Record<string, any>;
}

/**
 * Repositório para armazenar e recuperar templates
 */
export interface TemplateRepository {
  /**
   * Carrega um template pelo ID
   * @param id ID do template
   */
  getTemplate(id: string): Promise<TemplateModel | null>;
  
  /**
   * Lista templates disponíveis
   * @param filter Filtros opcionais
   */
  listTemplates(filter?: {
    categories?: string[];
    active?: boolean;
    search?: string;
  }): Promise<TemplateModel[]>;
  
  /**
   * Salva um template
   * @param template Template a ser salvo
   */
  saveTemplate(template: TemplateModel): Promise<TemplateModel>;
  
  /**
   * Remove um template
   * @param id ID do template
   */
  removeTemplate(id: string): Promise<boolean>;
}

/**
 * Repositório de templates baseado em sistema de arquivos
 */
export class FileSystemTemplateRepository implements TemplateRepository {
  private basePath: string;
  private logger = new Logger('FileSystemTemplateRepository');
  
  /**
   * Construtor
   * @param basePath Diretório base para os templates
   */
  constructor(basePath: string) {
    this.basePath = basePath;
    
    // Garantir que o diretório existe
    if (!fs.existsSync(this.basePath)) {
      fs.mkdirSync(this.basePath, { recursive: true });
      this.logger.info(`Diretório de templates criado: ${this.basePath}`);
    }
  }
  
  /**
   * Carrega um template pelo ID
   * @param id ID do template
   */
  async getTemplate(id: string): Promise<TemplateModel | null> {
    const filePath = path.join(this.basePath, `${id}.json`);
    
    try {
      if (!fs.existsSync(filePath)) {
        return null;
      }
      
      const fileContent = fs.readFileSync(filePath, 'utf8');
      const templateData = JSON.parse(fileContent);
      
      return this.deserializeTemplate(templateData);
    } catch (error) {
      this.logger.error(`Erro ao carregar template ${id}: ${error}`);
      return null;
    }
  }
  
  /**
   * Lista templates disponíveis
   * @param filter Filtros opcionais
   */
  async listTemplates(filter?: {
    categories?: string[];
    active?: boolean;
    search?: string;
  }): Promise<TemplateModel[]> {
    try {
      const files = fs.readdirSync(this.basePath).filter(f => f.endsWith('.json'));
      const templates: TemplateModel[] = [];
      
      for (const file of files) {
        try {
          const fileContent = fs.readFileSync(path.join(this.basePath, file), 'utf8');
          const templateData = JSON.parse(fileContent);
          const template = this.deserializeTemplate(templateData);
          
          // Aplicar filtros
          if (filter) {
            // Filtrar por categorias
            if (filter.categories && filter.categories.length > 0) {
              if (!template.categories || 
                  !filter.categories.some(c => template.categories!.includes(c))) {
                continue;
              }
            }
            
            // Filtrar por status ativo
            if (filter.active !== undefined && template.active !== filter.active) {
              continue;
            }
            
            // Filtrar por texto de busca
            if (filter.search) {
              const searchLower = filter.search.toLowerCase();
              const matches = 
                template.name.toLowerCase().includes(searchLower) ||
                template.description?.toLowerCase().includes(searchLower) ||
                template.categories?.some(c => c.toLowerCase().includes(searchLower));
              
              if (!matches) {
                continue;
              }
            }
          }
          
          templates.push(template);
        } catch (error) {
          this.logger.error(`Erro ao processar template ${file}: ${error}`);
        }
      }
      
      return templates;
    } catch (error) {
      this.logger.error(`Erro ao listar templates: ${error}`);
      return [];
    }
  }
  
  /**
   * Salva um template
   * @param template Template a ser salvo
   */
  async saveTemplate(template: TemplateModel): Promise<TemplateModel> {
    try {
      // Atualizar data de modificação
      template.updatedAt = new Date();
      
      const filePath = path.join(this.basePath, `${template.id}.json`);
      const serialized = this.serializeTemplate(template);
      
      fs.writeFileSync(filePath, JSON.stringify(serialized, null, 2));
      this.logger.info(`Template ${template.id} salvo com sucesso`);
      
      return template;
    } catch (error) {
      this.logger.error(`Erro ao salvar template ${template.id}: ${error}`);
      throw error;
    }
  }
  
  /**
   * Remove um template
   * @param id ID do template
   */
  async removeTemplate(id: string): Promise<boolean> {
    try {
      const filePath = path.join(this.basePath, `${id}.json`);
      
      if (!fs.existsSync(filePath)) {
        return false;
      }
      
      fs.unlinkSync(filePath);
      this.logger.info(`Template ${id} removido com sucesso`);
      
      return true;
    } catch (error) {
      this.logger.error(`Erro ao remover template ${id}: ${error}`);
      return false;
    }
  }
  
  /**
   * Serializa um template para armazenamento
   * @param template Template a ser serializado
   */
  private serializeTemplate(template: TemplateModel): any {
    return {
      ...template,
      channelTemplates: Object.fromEntries(template.channelTemplates),
      localizedData: Object.fromEntries(template.localizedData)
    };
  }
  
  /**
   * Deserializa um template do armazenamento
   * @param data Dados serializados
   */
  private deserializeTemplate(data: any): TemplateModel {
    return {
      ...data,
      channelTemplates: new Map(Object.entries(data.channelTemplates)),
      localizedData: new Map(Object.entries(data.localizedData)),
      createdAt: new Date(data.createdAt),
      updatedAt: new Date(data.updatedAt)
    };
  }
}

/**
 * Serviço para renderização de templates
 */
export class TemplateService {
  private repository: TemplateRepository;
  private templateCache: Map<string, TemplateModel> = new Map();
  private compiledTemplates: Map<string, handlebars.TemplateDelegate> = new Map();
  private logger = new Logger('TemplateService');
  
  /**
   * Construtor
   * @param repository Repositório de templates
   */
  constructor(repository: TemplateRepository) {
    this.repository = repository;
    this.registerHelpers();
  }
  
  /**
   * Registra helpers do Handlebars
   */
  private registerHelpers(): void {
    // Formatação de data
    handlebars.registerHelper('formatDate', (date, format) => {
      if (!date) return '';
      // Implementação básica - em produção usar bibliotecas como date-fns ou moment
      const d = new Date(date);
      return d.toLocaleDateString();
    });
    
    // Formatação de moeda
    handlebars.registerHelper('formatCurrency', (value, currency = 'USD', locale = 'en-US') => {
      if (value === undefined || value === null) return '';
      try {
        return new Intl.NumberFormat(locale, {
          style: 'currency',
          currency
        }).format(value);
      } catch (error) {
        return `${currency} ${value}`;
      }
    });
    
    // Truncamento de texto
    handlebars.registerHelper('truncate', (text, length) => {
      if (!text) return '';
      if (text.length <= length) return text;
      return text.substring(0, length) + '...';
    });
    
    // Transformação para maiúscula
    handlebars.registerHelper('uppercase', (text) => {
      if (!text) return '';
      return text.toUpperCase();
    });
    
    // Transformação para minúscula
    handlebars.registerHelper('lowercase', (text) => {
      if (!text) return '';
      return text.toLowerCase();
    });
    
    // Condicional se não vazio
    handlebars.registerHelper('ifNotEmpty', function(value, options) {
      if (!value || (Array.isArray(value) && value.length === 0) || value === '') {
        return options.inverse(this);
      }
      return options.fn(this);
    });
  }
  
  /**
   * Renderiza um template
   * @param templateId ID do template
   * @param options Opções de renderização
   */
  async render(
    templateId: string,
    options: TemplateRenderOptions
  ): Promise<RenderedTemplate> {
    // Carregar template (do cache ou repositório)
    let template = this.templateCache.get(templateId);
    
    if (!template) {
      template = await this.repository.getTemplate(templateId);
      
      if (!template) {
        throw new Error(`Template não encontrado: ${templateId}`);
      }
      
      if (!template.active) {
        throw new Error(`Template ${templateId} está inativo`);
      }
      
      // Armazenar em cache para uso futuro
      this.templateCache.set(templateId, template);
    }
    
    // Determinar o canal e verificar se o template suporta
    const channel = options.targetChannel;
    const channelTemplate = template.channelTemplates.get(channel);
    
    if (!channelTemplate) {
      throw new Error(`Template ${templateId} não suporta o canal ${channel}`);
    }
    
    // Determinar localidade a ser usada
    const locale = options.locale || 'pt-BR'; // Localidade padrão
    const localizedData = template.localizedData.get(locale) || template.defaultData;
    
    // Preparar dados para interpolação
    const templateData = {
      ...options.variables,
      _strings: localizedData.strings,
      _formats: localizedData.formats,
      _media: localizedData.mediaPreferences
    };
    
    // Renderizar componentes do template
    const subject = await this.renderTemplateContent(
      templateId,
      channelTemplate.subject || '',
      templateData,
      options,
      'subject'
    );
    
    const body = await this.renderTemplateContent(
      templateId,
      channelTemplate.body,
      templateData,
      options,
      'body'
    );
    
    let htmlBody: string | undefined;
    if (channelTemplate.htmlBody) {
      htmlBody = await this.renderTemplateContent(
        templateId,
        channelTemplate.htmlBody,
        templateData,
        options,
        'htmlBody'
      );
      
      // Adicionar pixel de rastreamento se solicitado
      if (options.transformOptions?.addTrackingPixel && 
          channel === NotificationChannel.EMAIL) {
        htmlBody = this.addTrackingPixel(htmlBody, options.variables);
      }
    } else if (options.transformOptions?.markdownToHtml && 
               channel === NotificationChannel.EMAIL && 
               body) {
      // Converter de markdown para HTML se solicitado
      htmlBody = await this.convertMarkdownToHtml(body);
    }
    
    // Componentes adicionais específicos do canal
    const components: Record<string, string> = {};
    if (channelTemplate.components) {
      for (const [key, template] of Object.entries(channelTemplate.components)) {
        components[key] = await this.renderTemplateContent(
          templateId,
          template,
          templateData,
          options,
          `component.${key}`
        );
      }
    }
    
    // Construir conteúdo da notificação
    const content: NotificationContent = {
      subject: subject || undefined,
      body,
      htmlBody,
      components
    };
    
    // Aplicar transformações finais específicas do canal
    if (options.transformOptions?.truncateToChannelLimits) {
      this.applyChannelLimits(content, channel);
    }
    
    return {
      content,
      templateId,
      templateVersion: template.version,
      channel,
      locale,
      variables: { ...options.variables }
    };
  }
  
  /**
   * Renderiza um conteúdo de template específico
   * @param templateId ID do template
   * @param content Conteúdo a ser renderizado
   * @param data Dados para interpolação
   * @param options Opções de renderização
   * @param partName Nome da parte para cache
   */
  private async renderTemplateContent(
    templateId: string,
    content: string,
    data: any,
    options: TemplateRenderOptions,
    partName: string
  ): Promise<string> {
    if (!content) {
      return '';
    }
    
    try {
      // Chave para cache de template compilado
      const templateKey = `${templateId}:${partName}`;
      
      // Obter ou compilar o template
      let compiledTemplate = this.compiledTemplates.get(templateKey);
      
      if (!compiledTemplate) {
        compiledTemplate = handlebars.compile(content);
        this.compiledTemplates.set(templateKey, compiledTemplate);
      }
      
      // Renderizar o template com os dados
      let rendered = compiledTemplate(data);
      
      // Aplicar transformações adicionais
      if (options.transformOptions?.trackUrls) {
        rendered = this.processUrlsForTracking(rendered, options.variables);
      }
      
      return rendered;
    } catch (error) {
      this.logger.error(`Erro ao renderizar template ${templateId} (${partName}): ${error}`);
      throw new Error(`Erro ao renderizar template: ${error.message}`);
    }
  }
  
  /**
   * Processa URLs no conteúdo para rastreamento
   * @param content Conteúdo com URLs
   * @param variables Variáveis de contexto
   */
  private processUrlsForTracking(content: string, variables: Record<string, any>): string {
    // Implementação básica - em produção usar expressões regulares mais robustas
    // e um serviço de URL shortener com rastreamento
    const urlRegex = /(https?:\/\/[^\s\)\"\']+)/g;
    
    return content.replace(urlRegex, (url) => {
      // Construir URL de rastreamento
      // Em produção, isso chamaria um serviço dedicado
      const trackingId = variables.notificationId || variables.recipientId || 'unknown';
      const separator = url.includes('?') ? '&' : '?';
      return `${url}${separator}utm_source=notification&utm_medium=email&utm_campaign=${variables.campaignId || 'general'}&tracking=${trackingId}`;
    });
  }
  
  /**
   * Adiciona pixel de rastreamento em conteúdo HTML
   * @param html Conteúdo HTML
   * @param variables Variáveis de contexto
   */
  private addTrackingPixel(html: string, variables: Record<string, any>): string {
    // Em produção, este seria um endpoint real de rastreamento
    const trackingId = variables.notificationId || variables.recipientId || 'unknown';
    const pixelUrl = `https://tracking.example.com/pixel?id=${trackingId}&t=${Date.now()}`;
    const trackingPixel = `<img src="${pixelUrl}" alt="" width="1" height="1" style="display:none;opacity:0;" />`;
    
    if (html.includes('</body>')) {
      return html.replace('</body>', `${trackingPixel}</body>`);
    } else {
      return `${html}${trackingPixel}`;
    }
  }
  
  /**
   * Converte conteúdo Markdown para HTML
   * @param markdown Conteúdo Markdown
   */
  private async convertMarkdownToHtml(markdown: string): Promise<string> {
    // Implementação básica - em produção usar biblioteca como marked
    // Esta é uma implementação muito simples para demonstração
    return markdown
      .replace(/\n\n/g, '</p><p>')
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      .replace(/\[(.*?)\]\((.*?)\)/g, '<a href="$2">$1</a>')
      .replace(/^# (.*)/gm, '<h1>$1</h1>')
      .replace(/^## (.*)/gm, '<h2>$1</h2>')
      .replace(/^### (.*)/gm, '<h3>$1</h3>')
      .replace(/^- (.*)/gm, '<li>$1</li>')
      .split(/<li>/).join('<ul><li>').split('</li>').join('</li></ul>')
      .replace(/<\/ul><ul>/g, '');
  }
  
  /**
   * Aplica limites específicos do canal ao conteúdo
   * @param content Conteúdo da notificação
   * @param channel Canal de notificação
   */
  private applyChannelLimits(content: NotificationContent, channel: NotificationChannel): void {
    switch (channel) {
      case NotificationChannel.SMS:
        // Limitar SMS para 160 caracteres
        if (content.body && content.body.length > 160) {
          content.body = content.body.substring(0, 157) + '...';
        }
        break;
        
      case NotificationChannel.PUSH:
        // Limitar título push para 50 caracteres e corpo para 200
        if (content.subject && content.subject.length > 50) {
          content.subject = content.subject.substring(0, 47) + '...';
        }
        if (content.body && content.body.length > 200) {
          content.body = content.body.substring(0, 197) + '...';
        }
        break;
        
      default:
        // Outros canais não têm limites rígidos ou são tratados pelos adaptadores
        break;
    }
  }
  
  /**
   * Limpa o cache de templates
   * @param templateId ID específico do template para limpar (opcional)
   */
  clearCache(templateId?: string): void {
    if (templateId) {
      this.templateCache.delete(templateId);
      
      // Limpar todas as partes compiladas deste template
      for (const key of this.compiledTemplates.keys()) {
        if (key.startsWith(`${templateId}:`)) {
          this.compiledTemplates.delete(key);
        }
      }
    } else {
      this.templateCache.clear();
      this.compiledTemplates.clear();
    }
  }
  
  /**
   * Cria um novo template
   * @param template Dados do template a ser criado
   */
  async createTemplate(template: Omit<TemplateModel, 'createdAt' | 'updatedAt'>): Promise<TemplateModel> {
    const now = new Date();
    const newTemplate: TemplateModel = {
      ...template,
      createdAt: now,
      updatedAt: now
    };
    
    return await this.repository.saveTemplate(newTemplate);
  }
  
  /**
   * Atualiza um template existente
   * @param templateId ID do template
   * @param updates Atualizações a serem aplicadas
   */
  async updateTemplate(
    templateId: string,
    updates: Partial<Omit<TemplateModel, 'id' | 'createdAt' | 'updatedAt'>>
  ): Promise<TemplateModel> {
    const template = await this.repository.getTemplate(templateId);
    
    if (!template) {
      throw new Error(`Template não encontrado: ${templateId}`);
    }
    
    // Aplicar atualizações
    Object.assign(template, updates);
    template.updatedAt = new Date();
    
    // Limpar cache
    this.clearCache(templateId);
    
    return await this.repository.saveTemplate(template);
  }
  
  /**
   * Obtém um template
   * @param templateId ID do template
   */
  async getTemplate(templateId: string): Promise<TemplateModel | null> {
    return this.repository.getTemplate(templateId);
  }
  
  /**
   * Lista templates disponíveis
   * @param filter Filtros opcionais
   */
  async listTemplates(filter?: {
    categories?: string[];
    active?: boolean;
    search?: string;
  }): Promise<TemplateModel[]> {
    return this.repository.listTemplates(filter);
  }
  
  /**
   * Remove um template
   * @param templateId ID do template
   */
  async removeTemplate(templateId: string): Promise<boolean> {
    this.clearCache(templateId);
    return this.repository.removeTemplate(templateId);
  }
}