/**
 * INNOVABIZ IAM - Componente de Validação
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação do componente de validação para o módulo Core IAM,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com segurança
 * total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.14.2 - Validação de dados)
 * - PCI DSS v4.0 (Requisito 6.2.4 - Validação de entrada)
 * - LGPD/GDPR/PDPA (Arts. 46 - Medidas de segurança)
 * - BNA Instrução 7/2021 (Art. 9 - Controles de validação)
 * - NIST CSF (PR.DS-2 - Proteção de dados em trânsito)
 * - OWASP ASVS 4.0 (V5 - Validação de entrada)
 */

package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/locales/pt_BR"
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	pt_translations "github.com/go-playground/validator/v10/translations/pt_BR"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/innovabiz/iam/internal/domain/errors"
)

// Validator define a interface para validação de dados
type Validator interface {
	// Validate valida um struct ou map e retorna os erros
	// Retorna um mapa com os campos e mensagens de erro
	Validate(data interface{}) (map[string]string, error)
	
	// ValidateField valida um único campo
	ValidateField(fieldName string, value interface{}, tag string) error
	
	// RegisterCustomValidation registra uma função de validação personalizada
	RegisterCustomValidation(tag string, fn validator.Func) error
}

// DefaultValidator implementa a interface Validator usando validator/v10
type DefaultValidator struct {
	validate   *validator.Validate
	translator ut.Translator
}

// NewValidator cria um novo validador com configurações padrão
func NewValidator() (*DefaultValidator, error) {
	// Criar instância do validador
	validate := validator.New()
	
	// Configurar para usar os nomes dos campos JSON ao invés dos nomes dos campos struct
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	// Configurar tradutor para português brasileiro
	ptBr := pt_BR.New()
	uni := ut.New(ptBr, ptBr)
	trans, found := uni.GetTranslator("pt_BR")
	if !found {
		return nil, fmt.Errorf("tradutor não encontrado")
	}
	
	// Registrar as traduções padrão
	if err := pt_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		return nil, fmt.Errorf("erro ao registrar traduções: %w", err)
	}
	
	// Criar validador
	v := &DefaultValidator{
		validate:   validate,
		translator: trans,
	}
	
	// Registrar validações personalizadas
	if err := v.registerCustomValidations(); err != nil {
		return nil, fmt.Errorf("erro ao registrar validações personalizadas: %w", err)
	}
	
	return v, nil
}

// Validate valida um struct ou map e retorna os erros
func (v *DefaultValidator) Validate(data interface{}) (map[string]string, error) {
	if data == nil {
		return nil, nil
	}
	
	// Validar os dados
	err := v.validate.Struct(data)
	if err == nil {
		return nil, nil
	}
	
	// Converter erros para um formato mais amigável
	validationErrors := make(map[string]string)
	
	if errors, ok := err.(validator.ValidationErrors); ok {
		// Erros de validação do validator
		for _, e := range errors {
			// Traduzir o erro
			validationErrors[e.Field()] = e.Translate(v.translator)
		}
	} else {
		// Outro tipo de erro
		return nil, fmt.Errorf("erro inesperado na validação: %w", err)
	}
	
	return validationErrors, nil
}

// ValidateField valida um único campo
func (v *DefaultValidator) ValidateField(fieldName string, value interface{}, tag string) error {
	err := v.validate.Var(value, tag)
	if err != nil {
		if errors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range errors {
				return fmt.Errorf("%s: %s", fieldName, e.Translate(v.translator))
			}
		}
		return fmt.Errorf("%s: erro de validação", fieldName)
	}
	return nil
}

// RegisterCustomValidation registra uma função de validação personalizada
func (v *DefaultValidator) RegisterCustomValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// registerCustomValidations registra todas as validações personalizadas
func (v *DefaultValidator) registerCustomValidations() error {
	// Validação de UUID
	if err := v.RegisterCustomValidation("uuid", validateUUID); err != nil {
		return err
	}
	
	// Validação de código de grupo
	if err := v.RegisterCustomValidation("groupcode", validateGroupCode); err != nil {
		return err
	}
	
	// Validação de JSON válido
	if err := v.RegisterCustomValidation("validjson", validateJSON); err != nil {
		return err
	}
	
	// Validação de tamanho máximo de JSON
	if err := v.RegisterCustomValidation("jsonmaxsize", validateJSONMaxSize); err != nil {
		return err
	}
	
	// Validação de caracteres seguros
	if err := v.RegisterCustomValidation("safetext", validateSafeText); err != nil {
		return err
	}
	
	return nil
}

// Funções de validação personalizadas

// validateUUID valida se uma string é um UUID válido
func validateUUID(fl validator.FieldLevel) bool {
	field := fl.Field()
	
	if field.Kind() != reflect.String {
		return false
	}
	
	str := field.String()
	if str == "" {
		return true // Permitir string vazia (campo opcional)
	}
	
	_, err := uuid.Parse(str)
	return err == nil
}

// validateGroupCode valida se um código de grupo está em formato válido
// Formato: apenas letras maiúsculas, números e underscore, começando com letra
func validateGroupCode(fl validator.FieldLevel) bool {
	field := fl.Field()
	
	if field.Kind() != reflect.String {
		return false
	}
	
	str := field.String()
	if str == "" {
		return true // Permitir string vazia (campo opcional)
	}
	
	// Padrão: apenas letras maiúsculas, números e underscore, começando com letra
	pattern := `^[A-Z][A-Z0-9_]{2,63}$`
	match, err := regexp.MatchString(pattern, str)
	return err == nil && match
}

// validateJSON valida se uma string é um JSON válido
func validateJSON(fl validator.FieldLevel) bool {
	field := fl.Field()
	
	if field.Kind() != reflect.String {
		return true // Não aplicável a tipos não-string
	}
	
	str := field.String()
	if str == "" {
		return true // Permitir string vazia
	}
	
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// validateJSONMaxSize valida se um JSON está dentro do tamanho máximo
func validateJSONMaxSize(fl validator.FieldLevel) bool {
	field := fl.Field()
	
	if field.Kind() != reflect.String {
		return true // Não aplicável a tipos não-string
	}
	
	str := field.String()
	if str == "" {
		return true // Permitir string vazia
	}
	
	// Tamanho máximo: 16KB
	maxSize := 16 * 1024
	return len(str) <= maxSize
}

// validateSafeText valida se uma string contém apenas caracteres seguros
func validateSafeText(fl validator.FieldLevel) bool {
	field := fl.Field()
	
	if field.Kind() != reflect.String {
		return true // Não aplicável a tipos não-string
	}
	
	str := field.String()
	if str == "" {
		return true // Permitir string vazia
	}
	
	// Verificar se todos os caracteres são UTF-8 válidos
	if !utf8.ValidString(str) {
		return false
	}
	
	// Lista de caracteres não permitidos
	dangerous := []string{
		"<script", "javascript:", "data:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
	}
	
	// Verificar se não contém sequências perigosas
	strLower := strings.ToLower(str)
	for _, pattern := range dangerous {
		if strings.Contains(strLower, pattern) {
			return false
		}
	}
	
	return true
}