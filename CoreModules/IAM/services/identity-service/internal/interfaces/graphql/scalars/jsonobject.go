/**
 * INNOVABIZ IAM - Scalar JSONObject para GraphQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação do scalar JSONObject para suporte a dados dinâmicos e flexíveis 
 * no módulo Core IAM, seguindo a arquitetura multi-dimensional, multi-tenant
 * e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (Gerenciamento flexível de identidades)
 * - PCI DSS v4.0 (Suporte a metadados e atributos customizados)
 * - ECMA-404 (Padrão JSON)
 * - RFC 8259 (Formato JSON)
 * - ISO/IEC 21778:2017 (JavaScript Object Notation)
 */

package scalars

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

// JSONObject representa um objeto JSON genérico para uso em GraphQL
type JSONObject map[string]interface{}

// MarshalJSON serializa um JSONObject para JSON
func (j JSONObject) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]interface{}(j))
}

// UnmarshalJSON deserializa um JSONObject a partir de JSON
func (j *JSONObject) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*j = nil
		return nil
	}
	
	m := make(map[string]interface{})
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	
	*j = JSONObject(m)
	return nil
}

// MarshalGQL implementa a interface Marshaler do GraphQL
func (j JSONObject) MarshalGQL(w io.Writer) {
	bytes, err := json.Marshal(j)
	if err != nil {
		w.Write([]byte(`null`))
		return
	}
	w.Write(bytes)
}

// UnmarshalGQL implementa a interface Unmarshaler do GraphQL
func (j *JSONObject) UnmarshalGQL(v interface{}) error {
	switch value := v.(type) {
	case map[string]interface{}:
		*j = JSONObject(value)
		return nil
	case JSONObject:
		*j = value
		return nil
	case nil:
		*j = nil
		return nil
	case string:
		// Tenta analisar string como JSON
		var m map[string]interface{}
		err := json.Unmarshal([]byte(value), &m)
		if err != nil {
			return fmt.Errorf("não foi possível analisar JSONObject da string: %v", err)
		}
		*j = JSONObject(m)
		return nil
	default:
		return fmt.Errorf("tipo inválido para JSONObject: %T", v)
	}
}

// JSONObjectScalar é o scalar personalizado para JSONObject no GraphQL
var JSONObjectScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "JSONObject",
	Description: "O scalar `JSONObject` representa um objeto JSON genérico como um map[string]interface{}",
	// Serializar o scalar para saída
	Serialize: func(value interface{}) interface{} {
		switch v := value.(type) {
		case JSONObject:
			return map[string]interface{}(v)
		case map[string]interface{}:
			return v
		case nil:
			return nil
		default:
			// Tentativa de conversão para JSON
			data, err := json.Marshal(v)
			if err != nil {
				return nil
			}
			var m map[string]interface{}
			err = json.Unmarshal(data, &m)
			if err != nil {
				return nil
			}
			return m
		}
	},
	// Analisar o valor como entrada de variável
	ParseValue: func(value interface{}) interface{} {
		switch v := value.(type) {
		case map[string]interface{}:
			return JSONObject(v)
		case JSONObject:
			return v
		case string:
			var m map[string]interface{}
			err := json.Unmarshal([]byte(v), &m)
			if err != nil {
				return nil
			}
			return JSONObject(m)
		case nil:
			return nil
		default:
			// Tentativa de conversão para JSON
			data, err := json.Marshal(v)
			if err != nil {
				return nil
			}
			var m map[string]interface{}
			err = json.Unmarshal(data, &m)
			if err != nil {
				return nil
			}
			return JSONObject(m)
		}
	},
	// Analisar o valor literal (para consultas GraphQL inline)
	ParseLiteral: func(valueAST interface{}) interface{} {
		switch v := valueAST.(type) {
		case *graphql.StringValue:
			var m map[string]interface{}
			err := json.Unmarshal([]byte(v.Value), &m)
			if err != nil {
				return nil
			}
			return JSONObject(m)
		case *graphql.ObjectValue:
			m := make(map[string]interface{})
			for _, field := range v.Fields {
				if value := parseASTLiteral(field.Value); value != nil {
					m[field.Name.Value] = value
				}
			}
			return JSONObject(m)
		default:
			return nil
		}
	},
})

// parseASTLiteral é uma função auxiliar para converter valores literais AST do GraphQL
func parseASTLiteral(valueAST interface{}) interface{} {
	switch v := valueAST.(type) {
	case *graphql.StringValue:
		return v.Value
	case *graphql.IntValue:
		i, err := strconv.ParseInt(v.Value, 10, 64)
		if err != nil {
			return nil
		}
		return i
	case *graphql.FloatValue:
		f, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			return nil
		}
		return f
	case *graphql.BooleanValue:
		return v.Value
	case *graphql.EnumValue:
		return v.Value
	case *graphql.ObjectValue:
		m := make(map[string]interface{})
		for _, field := range v.Fields {
			if value := parseASTLiteral(field.Value); value != nil {
				m[field.Name.Value] = value
			}
		}
		return m
	case *graphql.ListValue:
		var list []interface{}
		for _, item := range v.Values {
			if value := parseASTLiteral(item); value != nil {
				list = append(list, value)
			}
		}
		return list
	default:
		return nil
	}
}