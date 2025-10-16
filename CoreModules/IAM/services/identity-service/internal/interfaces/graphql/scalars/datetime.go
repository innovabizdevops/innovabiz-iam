/**
 * INNOVABIZ IAM - Scalar DateTime para GraphQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação do scalar DateTime para integração com GraphQL no módulo Core IAM,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade
 * total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO 8601:2004 (Representação de data e hora)
 * - RFC 3339 (Timestamp em Internet)
 * - ISO/IEC 27001:2022 (Rastreabilidade de dados)
 */

package scalars

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// DateTime é uma implementação personalizada do scalar DateTime para GraphQL
type DateTime struct {
	time.Time
}

// MarshalJSON serializa um DateTime para JSON
func (t DateTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", t.Time.Format(time.RFC3339))
	return []byte(stamp), nil
}

// UnmarshalJSON deserializa um DateTime a partir de JSON
func (t *DateTime) UnmarshalJSON(data []byte) error {
	var err error
	var v interface{}
	
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	
	switch value := v.(type) {
	case string:
		t.Time, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		return nil
	case float64:
		t.Time = time.Unix(int64(value), 0)
		return nil
	case map[string]interface{}:
		// Manipular caso de objeto Date
		if timeStr, ok := value["time"].(string); ok {
			t.Time, err = time.Parse(time.RFC3339, timeStr)
			return err
		}
		return fmt.Errorf("não foi possível analisar objeto DateTime: %v", value)
	default:
		return fmt.Errorf("tipo inválido para DateTime: %T", v)
	}
}

// MarshalGQL implementa a interface Marshaler do GraphQL
func (t DateTime) MarshalGQL(w io.Writer) {
	w.Write([]byte(strconv.Quote(t.Time.Format(time.RFC3339))))
}

// UnmarshalGQL implementa a interface Unmarshaler do GraphQL
func (t *DateTime) UnmarshalGQL(v interface{}) error {
	switch value := v.(type) {
	case string:
		var err error
		t.Time, err = time.Parse(time.RFC3339, value)
		return err
	case time.Time:
		t.Time = value
		return nil
	case *time.Time:
		if value != nil {
			t.Time = *value
		}
		return nil
	case int:
		t.Time = time.Unix(int64(value), 0)
		return nil
	case int64:
		t.Time = time.Unix(value, 0)
		return nil
	case float64:
		t.Time = time.Unix(int64(value), 0)
		return nil
	default:
		return fmt.Errorf("tipo de entrada inválido para DateTime: %T", v)
	}
}

// DateTimeScalar é o scalar personalizado para DateTime no GraphQL
var DateTimeScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "DateTime",
	Description: "O scalar `DateTime` representa um valor de data e hora ISO-8601 (RFC 3339)",
	// Serializar o scalar para saída
	Serialize: func(value interface{}) interface{} {
		switch v := value.(type) {
		case DateTime:
			return v.Time.Format(time.RFC3339)
		case *DateTime:
			if v == nil {
				return nil
			}
			return v.Time.Format(time.RFC3339)
		case time.Time:
			return v.Format(time.RFC3339)
		case *time.Time:
			if v == nil {
				return nil
			}
			return v.Format(time.RFC3339)
		default:
			return fmt.Sprintf("tipo inválido para DateTime: %T", value)
		}
	},
	// Analisar o valor como entrada de variável
	ParseValue: func(value interface{}) interface{} {
		switch v := value.(type) {
		case string:
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return nil
			}
			return DateTime{Time: t}
		case time.Time:
			return DateTime{Time: v}
		case *time.Time:
			if v == nil {
				return nil
			}
			return DateTime{Time: *v}
		case int:
			return DateTime{Time: time.Unix(int64(v), 0)}
		case int64:
			return DateTime{Time: time.Unix(v, 0)}
		case float64:
			return DateTime{Time: time.Unix(int64(v), 0)}
		default:
			return nil
		}
	},
	// Analisar o valor literal (para consultas GraphQL inline)
	ParseLiteral: func(valueAST interface{}) interface{} {
		switch v := valueAST.(type) {
		case *graphql.StringValue:
			t, err := time.Parse(time.RFC3339, v.Value)
			if err != nil {
				return nil
			}
			return DateTime{Time: t}
		case *graphql.IntValue:
			n, err := strconv.ParseInt(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			return DateTime{Time: time.Unix(n, 0)}
		default:
			return nil
		}
	},
})