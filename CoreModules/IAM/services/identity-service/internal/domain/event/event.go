/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Definição da interface base para eventos no sistema IAM.
 * Implementa o contrato fundamental para todos os eventos do domínio.
 * Segue princípios de Event-Driven Architecture e Domain-Driven Design (DDD).
 */

package event

import (
	"context"
	"time"
)

// Event interface base para todos os eventos do domínio
type Event interface {
	GetType() string
	GetTime() time.Time
}

// EventBus interface para publicação e assinatura de eventos
type EventBus interface {
	Publish(ctx context.Context, eventType string, event Event) error
	Subscribe(eventType string, handler func(ctx context.Context, event Event) error) error
	Unsubscribe(eventType string, handler func(ctx context.Context, event Event) error) error
}