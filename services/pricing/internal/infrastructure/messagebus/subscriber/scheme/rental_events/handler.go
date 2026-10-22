package rental_events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/gofrs/uuid"

	"pricing/internal/pkg/inbox/message"
)

type messageSaver interface {
	SaveMessages(context.Context, message.Messages) error
}

type MessageHandler struct {
	inbox messageSaver
}

func NewMessageHandler(inbox messageSaver) *MessageHandler {
	return &MessageHandler{inbox: inbox}
}

func (h *MessageHandler) Handle(ctx context.Context, _ sarama.ConsumerGroupSession, kafkaMessage *sarama.ConsumerMessage) error {
	var base BaseEvent
	if err := json.Unmarshal(kafkaMessage.Value, &base); err != nil {
		return err
	}
	headers := make(map[string]string, len(kafkaMessage.Headers))
	for _, header := range kafkaMessage.Headers {
		headers[string(header.Key)] = string(header.Value)
	}
	correlationID := headers[correlationIDHeader]
	if uuid.FromStringOrNil(correlationID) == uuid.Nil {
		return fmt.Errorf("в событии отсутствует корректный x-correlation-id")
	}
	if correlationID != base.CorrelationID {
		return fmt.Errorf("correlation-id в заголовке и событии не совпадают")
	}
	if uuid.FromStringOrNil(base.RentalID) == uuid.Nil || string(kafkaMessage.Key) != base.RentalID {
		return fmt.Errorf("ключ Kafka не соответствует rental_id")
	}
	if headers["x-event-type"] != base.EventType || base.EventType == "" {
		return fmt.Errorf("тип события в заголовке и envelope не совпадает")
	}
	headersRaw, err := json.Marshal(headers)
	if err != nil {
		return err
	}
	msg := message.NewMessage(
		base.RentalID,
		correlationID,
		kafkaMessage.Topic,
		kafkaMessage.Key,
		kafkaMessage.Value,
		kafkaMessage.Timestamp,
		message.WithHeaders(headersRaw),
	)
	return h.inbox.SaveMessages(ctx, message.Messages{msg})
}
