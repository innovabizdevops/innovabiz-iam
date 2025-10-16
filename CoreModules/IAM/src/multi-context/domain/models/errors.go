/**
 * @file errors.go
 * @description Definição de erros comuns para o domínio de identidade multi-contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package models

import "errors"

// Erros comuns do domínio de identidade
var (
	// ErrInvalidPrimaryKey ocorre quando uma chave primária é inválida
	ErrInvalidPrimaryKey = errors.New("chave primária inválida")
	
	// ErrIdentityNotFound ocorre quando uma identidade não é encontrada
	ErrIdentityNotFound = errors.New("identidade não encontrada")
	
	// ErrContextNotFound ocorre quando um contexto específico não é encontrado
	ErrContextNotFound = errors.New("contexto não encontrado")
	
	// ErrDuplicateIdentity ocorre quando há tentativa de criar identidade duplicada
	ErrDuplicateIdentity = errors.New("identidade já existe")
	
	// ErrDuplicateContext ocorre quando há tentativa de criar contexto duplicado
	ErrDuplicateContext = errors.New("contexto já existe para esta identidade")
	
	// ErrInvalidContextType ocorre quando o tipo de contexto é inválido
	ErrInvalidContextType = errors.New("tipo de contexto inválido")
	
	// ErrInvalidAttribute ocorre quando um atributo é inválido
	ErrInvalidAttribute = errors.New("atributo inválido")
	
	// ErrIdentityLocked ocorre quando a identidade está bloqueada
	ErrIdentityLocked = errors.New("identidade bloqueada")
	
	// ErrUnauthorizedContextAccess ocorre quando acesso a um contexto é não autorizado
	ErrUnauthorizedContextAccess = errors.New("acesso não autorizado ao contexto")
	
	// ErrMissingConsent ocorre quando está faltando consentimento para acesso
	ErrMissingConsent = errors.New("consentimento necessário para acesso ao contexto")
)