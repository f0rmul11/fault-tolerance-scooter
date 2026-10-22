package update

import (
	"github.com/gofrs/uuid"
	"github.com/samber/mo"
)

type Input struct {
	ID              uuid.UUID
	ExpectedVersion int
	Number          mo.Option[string]
	Brand           mo.Option[string]
	ChargePercent   mo.Option[int]
	RiddenSeconds   mo.Option[int64]
}
