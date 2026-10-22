package scooter

import "github.com/gofrs/uuid"

type Scooter struct {
	ID            uuid.UUID
	Number        string
	Brand         string
	ChargePercent int
}
