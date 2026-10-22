package producer

import (
	"rental/internal/pkg/msgbus/producer"
)

type Producers struct {
	RentalEvents *producer.MessageProducer
}

func NewProducers() Producers {
	return Producers{
		RentalEvents: producer.NewMessageProducer(),
	}
}
