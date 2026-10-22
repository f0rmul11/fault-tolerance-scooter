package subscriber

import (
	"pricing/config"
	"pricing/internal/pkg/msgbus/subscriber"
)

type Subscribers struct{ RentalEvents *subscriber.MessageSubscriber }

func NewSubscribers() Subscribers {
	return Subscribers{RentalEvents: subscriber.NewMessageSubscriber(config.RentalEventsTopic)}
}
