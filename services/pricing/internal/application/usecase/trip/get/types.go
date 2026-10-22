package get

import "github.com/gofrs/uuid"

type Input struct {
	RentalID uuid.UUID
	UserID   uuid.UUID
}
