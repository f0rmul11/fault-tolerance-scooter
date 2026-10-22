package messagebus

import (
	"rental/internal/infrastructure/messagebus/producer"
	"rental/internal/pkg/closer"
)

type Registry struct {
	Producers producer.Producers
}

func NewRegistry() *Registry {
	registry := &Registry{
		Producers: producer.NewProducers(),
	}
	closer.Add(registry.Producers.RentalEvents.Close)
	return registry
}
