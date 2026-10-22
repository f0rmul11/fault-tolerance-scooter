package update

import (
	"github.com/gofrs/uuid"
	"github.com/samber/mo"

	"profile/internal/domain/entity"
)

type Input struct {
	ID              uuid.UUID
	ExpectedVersion int
	Name            mo.Option[string]
	Email           mo.Option[string]
	Plan            mo.Option[entity.Plan]
}
