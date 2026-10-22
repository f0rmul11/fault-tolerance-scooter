package list

import "github.com/gofrs/uuid"

type Input struct {
	UserID uuid.UUID
	Limit  int
	Offset int
}
