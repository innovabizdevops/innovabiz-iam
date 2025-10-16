module github.com/innovabiz/iam/services/identity-service

go 1.21

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/go-chi/jwtauth/v5 v5.1.1
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/graph-gophers/graphql-go v1.5.0
	github.com/jackc/pgx/v5 v5.4.3
	github.com/jmoiron/sqlx v1.3.5
	github.com/lestrrat-go/jwx/v2 v2.0.11
	github.com/prometheus/client_golang v1.16.0
	github.com/redis/go-redis/v9 v9.1.0
	github.com/rs/zerolog v1.30.0
	github.com/spf13/viper v1.16.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.42.0
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.16.0
	go.opentelemetry.io/otel/sdk v1.16.0
	golang.org/x/crypto v0.13.0
	golang.org/x/sync v0.3.0
	google.golang.org/grpc v1.58.1
)