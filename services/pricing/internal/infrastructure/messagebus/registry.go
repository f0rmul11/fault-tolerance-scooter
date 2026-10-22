package messagebus

import (
	"context"
	"pricing/internal/infrastructure/messagebus/subscriber"
	"pricing/internal/infrastructure/messagebus/subscriber/scheme/rental_events"
	"pricing/internal/pkg/closer"
	"pricing/internal/pkg/inbox"
)

type handlers struct {
	RentalEvents *rental_events.MessageHandler
}

type Registry struct {
	handlers    handlers
	subscribers subscriber.Subscribers
}

func NewRegistry(inbox *inbox.Inbox) *Registry {
	registry := &Registry{
		subscribers: subscriber.NewSubscribers(),
		handlers:    handlers{RentalEvents: rental_events.NewMessageHandler(inbox)},
	}
	closer.Add(registry.subscribers.RentalEvents.Close)
	return registry
}

func (r *Registry) Run(ctx context.Context) {
	r.subscribers.RentalEvents.Subscribe(ctx, r.handlers.RentalEvents.Handle)
}

func (r *Registry) Stop() error {
	r.subscribers.RentalEvents.Stop()
	return nil
}
