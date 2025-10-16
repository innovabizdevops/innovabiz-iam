/**
 * INNOVABIZ IAM - Diretiva de Validação GraphQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação da diretiva de validação de entrada para GraphQL
 * no módulo Core IAM, seguindo a arquitetura multi-dimensional, multi-tenant
 * e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Validação de entrada)
 * - PCI DSS v4.0 (Requisito 6.2.4 - Validação de entrada)
 * - LGPD/GDPR/PDPA (Arts. 46 - Medidas técnicas adequadas)
 * - BNA Instrução 7/2021 (Art. 9 - Controles de validação)
 * - NIST CSF (PR.DS-2 - Proteção de dados em trânsito)
 */

package directives

import (
	"context"
	"reflect"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"go.opentelemetry.io/otel/attribute"

	"github.com/innovabiz/iam/internal/domain/errors"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/metrics"
	"github.com/innovabiz/iam/internal/infrastructure/tracing"
	"github.com/innovabiz/iam/internal/infrastructure/validation"
)

// ValidateInputDirective implementa a diretiva @validateInput para validação de entradas GraphQL
func ValidateInputDirective(
	logger logging.Logger,
	metrics metrics.MetricsClient,
	tracer tracing.Tracer,
	validator validation.Validator,
) func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		span, ctx := tracer.StartFromContext(ctx, "ValidateInputDirective")
		defer span.End()

		operationName := graphql.GetOperationContext(ctx).OperationName
		if operationName == "" {
			operationName = "unknown"
		}

		logger := logger.WithContext(ctx).
			WithField("directive", "validateInput").
			WithField("operation", operationName)

		// Obter o campo que está sendo resolvido
		fieldContext := graphql.GetFieldContext(ctx)
		inputName := fieldContext.Field.Name

		span.SetAttributes(attribute.String("input_field", inputName))

		// Obter argumentos da operação
		args := fieldContext.Field.Arguments
		if len(args) == 0 {
			// Sem argumentos para validar
			return next(ctx)
		}

		logger.WithField("input_name", inputName).Info("Validando entrada")

		// Validar cada argumento
		for _, arg := range args {
			argValue := arg.Value.Value(ctx)
			if argValue == nil {
				continue
			}

			argName := arg.Name
			span.SetAttributes(attribute.String("arg_name", argName))

			// Validar apenas structs e maps
			v := reflect.ValueOf(argValue)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			// Ignorar tipos primitivos
			if v.Kind() != reflect.Struct && v.Kind() != reflect.Map {
				continue
			}

			logger.WithField("arg_name", argName).Debug("Validando argumento")

			// Realizar a validação
			validationErrs, err := validator.Validate(argValue)
			if err != nil {
				logger.WithError(err).Error("Erro interno ao validar entrada")
				span.RecordError(err)
				metrics.IncValidationErrors("internal_error", inputName)
				return nil, errors.NewInternalError("Erro ao validar entrada")
			}

			if len(validationErrs) > 0 {
				// Construir mensagem de erro detalhada
				errorMessages := make([]string, 0, len(validationErrs))
				for field, errMsg := range validationErrs {
					errorMessages = append(errorMessages, field+": "+errMsg)
				}

				errorMsg := strings.Join(errorMessages, "; ")
				logger.WithField("validation_errors", errorMessages).Warn("Falha na validação")
				
				span.SetAttributes(attribute.StringSlice("validation_errors", errorMessages))
				metrics.IncValidationErrors("validation_failed", inputName)
				
				return nil, errors.NewValidationError(errorMsg)
			}
		}

		// Registrar validação bem-sucedida
		metrics.IncValidationSuccess(inputName)
		logger.Info("Validação bem-sucedida")

		return next(ctx)
	}
}

// MockValidateInputDirective é uma versão simplificada da diretiva para ambientes de desenvolvimento
func MockValidateInputDirective(
	logger logging.Logger,
	metrics metrics.MetricsClient,
	tracer tracing.Tracer,
) func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		span, ctx := tracer.StartFromContext(ctx, "MockValidateInputDirective")
		defer span.End()

		operationName := graphql.GetOperationContext(ctx).OperationName
		logger := logger.WithContext(ctx).
			WithField("directive", "mock_validate_input").
			WithField("operation", operationName)

		logger.Info("Usando diretiva de validação simulada para ambiente de desenvolvimento")
		
		// Sem validação real, apenas para testes
		return next(ctx)
	}
}