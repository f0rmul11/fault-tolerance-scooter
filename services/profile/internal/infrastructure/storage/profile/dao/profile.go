package dao

import (
	"database/sql"

	"profile/internal/domain/entity"
	"profile/internal/pkg/xo"
)

type User struct {
	xo.User
	Plan sql.NullString
}

func (row User) ConvertTo() *entity.User {
	return entity.With(
		row.ID,
		entity.WithName(row.Name.String),
		entity.WithEmail(row.Email.String),
		entity.WithUpdatedAt(row.UpdatedAt),
		entity.WithPlan(entity.Plan(row.Plan.String)),
		entity.WithVersion(int(row.Version.Int64)),
	)
}
