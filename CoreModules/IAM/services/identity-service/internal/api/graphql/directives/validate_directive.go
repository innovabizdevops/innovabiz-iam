package directives

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/innovabiz/iam/internal/domain/model/errors"
	"github.com/innovabiz/iam/internal/infrastructure/observability"
	"github.com/innovabiz/iam/internal/infrastructure/validation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ValidateDirective implementa a diretiva @validateInput do GraphQL
// para validação automática de inputs usando o validator customizado.
type ValidateDirective struct {
	validator *validation.Validator
	tracer    trace.Tracer
	logger    observability.Logger
}

// NewValidateDirective cria uma nova instância da diretiva de validação
func NewValidateDirective(validator *validation.Validator, tracer trace.Tracer, logger observability.Logger) *ValidateDirective {
	return &ValidateDirective{
		validator: validator,
		tracer:    tracer,
		logger:    logger,
	}
}

// Directive é o ponto de entrada para a diretiva @validateInput e será chamada pelo
// runtime do GraphQL para validar inputs antes de executar o resolver
func (d *ValidateDirective) Directive() func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		// Iniciar span para observabilidade da validação
		validationCtx, span := d.tracer.Start(ctx, "validation.directive.validate")
		defer span.End()

		// Obter o campo atual que está sendo resolvido
		fieldContext := graphql.GetFieldContext(ctx)
		if fieldContext == nil {
			d.logger.Warn(ctx, "Validate directive: no field context available")
			span.SetAttributes(attribute.Bool("validation.success", false))
			span.SetAttributes(attribute.String("validation.error", "no_field_context"))
			return nil, errors.NewInternalError("validation_error", "Erro interno na validação")
		}

		// Logging da operação de validação
		d.logger.Debug(ctx, "Validating input",
			"field", fieldContext.Field.Name,
			"parent_type", fieldContext.Object,
		)

		// Coletar os argumentos a serem validados
		inputArgs, inputsToValidate := d.collectInputsToValidate(fieldContext)
		if len(inputsToValidate) == 0 {
			// Se não há inputs para validar, prossiga normalmente
			span.SetAttributes(attribute.Bool("validation.success", true))
			span.SetAttributes(attribute.Int("validation.input_count", 0))
			return next(ctx)
		}

		span.SetAttributes(attribute.Int("validation.input_count", len(inputsToValidate)))

		// Validar cada input
		for argName, input := range inputsToValidate {
			// Realizar validação do input usando o validator customizado
			err := d.validator.Validate(input)
			if err != nil {
				// Converter erro de validação para formato GraphQL amigável
				validationErrors := d.formatValidationErrors(err)
				
				d.logger.Warn(ctx, "Input validation failed",
					"field", fieldContext.Field.Name,
					"arg", argName,
					"errors", validationErrors,
				)

				span.SetAttributes(attribute.Bool("validation.success", false))
				span.SetAttributes(attribute.String("validation.failed_input", argName))
				
				// Retornar erro de validação formatado para o cliente
				return nil, errors.NewValidationError(validationErrors)
			}

			// Adicionar metadados ao span para cada input validado com sucesso
			span.SetAttributes(attribute.String(fmt.Sprintf("validation.input.%s", argName), "valid"))
		}

		// Logging de validação bem-sucedida
		d.logger.Debug(ctx, "Input validation successful",
			"field", fieldContext.Field.Name,
			"validated_inputs", inputArgs,
		)

		// Marcar validação como bem-sucedida no span
		span.SetAttributes(attribute.Bool("validation.success", true))

		// Prosseguir para o resolver com todos os inputs validados
		return next(validationCtx)
	}
}

// collectInputsToValidate extrai os argumentos que precisam ser validados
func (d *ValidateDirective) collectInputsToValidate(fieldContext *graphql.FieldContext) ([]string, map[string]interface{}) {
	inputsToValidate := make(map[string]interface{})
	var inputArgs []string

	// Iterar pelos argumentos do campo GraphQL
	for argName, argValue := range fieldContext.Args {
		// Verificar se o argumento é um objeto complexo que precisa de validação
		// (ignoramos tipos primitivos como string, int, etc.)
		if d.isValidatableInput(argValue) {
			inputsToValidate[argName] = argValue
			inputArgs = append(inputArgs, argName)
		}
	}

	return inputArgs, inputsToValidate
}

// isValidatableInput verifica se um valor é um tipo que deve ser validado
func (d *ValidateDirective) isValidatableInput(value interface{}) bool {
	// Ignorar valores nulos
	if value == nil {
		return false
	}

	// Verificar o tipo do valor
	switch reflect.TypeOf(value).Kind() {
	case reflect.Struct, reflect.Map:
		// Estruturas e mapas são candidatos a validação
		return true
	case reflect.Ptr:
		// Para ponteiros, verificar o tipo subjacente
		ptrValue := reflect.ValueOf(value)
		if ptrValue.IsNil() {
			return false
		}
		return d.isValidatableInput(ptrValue.Elem().Interface())
	case reflect.Slice, reflect.Array:
		// Para slices e arrays, verificamos se contêm objetos complexos
		sliceValue := reflect.ValueOf(value)
		if sliceValue.Len() > 0 {
			// Verificar o primeiro elemento como amostra
			return d.isValidatableInput(sliceValue.Index(0).Interface())
		}
	}

	// Por padrão, tipos primitivos não precisam de validação estruturada
	return false
}

// formatValidationErrors converte erros de validação para um formato amigável ao cliente GraphQL
func (d *ValidateDirective) formatValidationErrors(err error) map[string][]string {
	if validationErr, ok := err.(*validation.ValidationError); ok {
		return validationErr.Errors
	}

	// Para outros tipos de erro, criar um formato genérico
	return map[string][]string{
		"general": {err.Error()},
	}
}

// customValidationRules define regras de validação específicas para campos GraphQL
func (d *ValidateDirective) customValidationRules() map[string]map[string]string {
	// Mapeamento de campos GraphQL para regras de validação específicas
	// Formato: mutation.arg.field: "regra"
	return map[string]map[string]string{
		"createUser": {
			"input.username": "required,min=3,max=50,alphanum",
			"input.email": "required,email",
			"input.firstName": "required,max=100",
			"input.lastName": "required,max=100",
			"input.password": "required,min=8,containsAny=!@#$%^&*",
		},
		"updateUser": {
			"input.email": "omitempty,email",
			"input.firstName": "omitempty,max=100",
			"input.lastName": "omitempty,max=100",
		},
		"createGroup": {
			"input.name": "required,min=2,max=100",
			"input.code": "required,min=2,max=50,alphanum",
		},
		// Adicionar outras regras específicas aqui
	}
}

// getFieldValidationRules obtém regras de validação específicas para um campo GraphQL
func (d *ValidateDirective) getFieldValidationRules(fieldContext *graphql.FieldContext, argName string, path string) string {
	// Construir o caminho completo para buscar as regras
	operationName := fieldContext.Field.Name
	fullPath := fmt.Sprintf("%s.%s%s", operationName, argName, path)

	// Buscar nas regras customizadas
	rules := d.customValidationRules()
	if fieldRules, exists := rules[operationName]; exists {
		if rule, exists := fieldRules[fullPath]; exists {
			return rule
		}
	}

	// Retornar regras padrão se não houver específicas
	return ""
}