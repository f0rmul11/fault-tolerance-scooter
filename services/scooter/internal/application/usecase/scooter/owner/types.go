package owner

import "github.com/gofrs/uuid"

// Input содержит новое значение одного поля и условие атомарной записи.
type Input struct {
	ID              uuid.UUID
	ExpectedVersion int
	Owner           string
}
