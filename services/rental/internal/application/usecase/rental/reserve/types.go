package reserve

import "github.com/gofrs/uuid"

type Input struct {
	UserID         uuid.UUID
	ScooterID      uuid.UUID
	IdempotencyKey uuid.UUID
	RequestID      uuid.UUID
}
