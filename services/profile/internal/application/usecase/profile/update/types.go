package update

import (
	"profile/internal/domain/entity"

	"github.com/gofrs/uuid"
	"github.com/samber/mo"
)

type Input struct {
	ID              uuid.UUID
	ExpectedVersion int
	Name            mo.Option[string]
	Email           mo.Option[string]
	Plan            mo.Option[entity.Plan]
}
