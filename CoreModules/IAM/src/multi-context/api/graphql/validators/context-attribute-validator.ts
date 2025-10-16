import { 
  CreateContextAttributeInput, 
  UpdateContextAttributeInput,
  SensitivityLevel 
} from '../types/generated';
import { ContextAttributeRepository } from '../../../repositories/context-attribute-repository';
import { ContextRepository } from '../../../repositories/context-repository';
import { ValidationError } from '../../../../infrastructure/common/errors/validation-error';

/**
 * Validador para operações relacionadas a atributos de contexto
 */
export class ContextAttributeValidator {
  private attributeRepository: ContextAttributeRepository;
  private contextRepository: ContextRepository;
  
  constructor() {
    this.attributeRepository = new ContextAttributeRepository();
    this.contextRepository = new ContextRepository();
  }
  
  /**
   * Valida dados para criação de um novo atributo de contexto
   * @param input - Dados de entrada para criação
   */
  async validateCreate(input: CreateContextAttributeInput): Promise<void> {
    const errors: string[] = [];
    
    // Verifica se o contextoID é válido
    if (!input.contextId) {
      errors.push('O ID do contexto é obrigatório');
    } else {
      const contextExists = await this.contextRepository.exists(input.contextId);
      if (!contextExists) {
        errors.push(`Contexto não encontrado com ID: ${input.contextId}`);
      }
    }
    
    // Validação da chave do atributo
    if (!input.attributeKey) {
      errors.push('A chave do atributo é obrigatória');
    } else if (input.attributeKey.length < 2) {
      errors.push('A chave do atributo deve ter pelo menos 2 caracteres');
    } else if (input.attributeKey.length > 100) {
      errors.push('A chave do atributo não pode ter mais de 100 caracteres');
    } else if (!/^[a-zA-Z0-9_.:-]+$/.test(input.attributeKey)) {
      errors.push('A chave do atributo deve conter apenas letras, números, pontos, underscores, hífens e dois-pontos');
    }
    
    // Verifica se já existe um atributo com a mesma chave
    if (input.contextId && input.attributeKey) {
      const existingAttribute = await this.attributeRepository.findByContextAndKey(
        input.contextId,
        input.attributeKey
      );
      
      if (existingAttribute) {
        errors.push(`Já existe um atributo com a chave '${input.attributeKey}' neste contexto`);
      }
    }
    
    // Validação do nível de sensibilidade
    if (input.sensitivityLevel) {
      const validLevels: SensitivityLevel[] = ['PUBLIC', 'INTERNAL', 'CONFIDENTIAL', 'RESTRICTED'];
      if (!validLevels.includes(input.sensitivityLevel)) {
        errors.push(`Nível de sensibilidade inválido: ${input.sensitivityLevel}`);
      }
    }
    
    // Validação de metadados
    if (input.metadata && typeof input.metadata === 'object') {
      try {
        // Verifica se o objeto de metadados é serializável
        JSON.stringify(input.metadata);
      } catch (error) {
        errors.push('Metadados inválidos: deve ser um objeto JSON serializável');
      }
    }
    
    // Lança erro se houver validações que falharam
    if (errors.length > 0) {
      throw new ValidationError('Erro na validação de atributo de contexto', errors);
    }
  }
  
  /**
   * Valida dados para atualização de um atributo de contexto existente
   * @param input - Dados de entrada para atualização
   */
  async validateUpdate(input: UpdateContextAttributeInput): Promise<void> {
    const errors: string[] = [];
    
    // Verifica se o ID do atributo é válido
    if (!input.attributeId) {
      errors.push('O ID do atributo é obrigatório');
    } else {
      const attribute = await this.attributeRepository.getById(input.attributeId);
      if (!attribute) {
        errors.push(`Atributo não encontrado com ID: ${input.attributeId}`);
      } else {
        // Se o atributo existir e não for mutável, verifica se está tentando alterar o valor
        if (!attribute.isMutable && 
            input.attributeValue !== undefined && 
            input.attributeValue !== attribute.attributeValue) {
          errors.push(`O atributo '${attribute.attributeKey}' não é mutável e seu valor não pode ser alterado`);
        }
      }
    }
    
    // Validação do nível de sensibilidade
    if (input.sensitivityLevel) {
      const validLevels: SensitivityLevel[] = ['PUBLIC', 'INTERNAL', 'CONFIDENTIAL', 'RESTRICTED'];
      if (!validLevels.includes(input.sensitivityLevel)) {
        errors.push(`Nível de sensibilidade inválido: ${input.sensitivityLevel}`);
      }
    }
    
    // Validação de metadados
    if (input.metadata && typeof input.metadata === 'object') {
      try {
        // Verifica se o objeto de metadados é serializável
        JSON.stringify(input.metadata);
      } catch (error) {
        errors.push('Metadados inválidos: deve ser um objeto JSON serializável');
      }
    }
    
    // Lança erro se houver validações que falharam
    if (errors.length > 0) {
      throw new ValidationError('Erro na validação de atributo de contexto', errors);
    }
  }
}