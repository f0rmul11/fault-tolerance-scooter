package profile

import (
	"context"
	"errors"
	"profile/internal/domain/entity"
	"profile/internal/infrastructure/storage/profile/dao"
	app "profile/internal/pkg/profileerror"
	"profile/internal/pkg/transaction/wrapper"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
)

type Storage struct {
	db wrapper.Database
}

func New(db wrapper.Database) *Storage {
	return &Storage{db: db}
}
func (s *Storage) Get(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	const query = `
	select u.*, s.plan
	from public.users u
		join public.subscriptions s on s.user_id = u.id
	where u.id = $1
	`
	var row dao.User
	if err := pgxscan.Get(ctx, s.db.Pool(ctx), &row, query, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app.NewNotFoundErr("Запись не найдена", err)
		}
		return nil, err
	}
	return row.ConvertTo(), nil
}
func (s *Storage) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	const query = `
	select u.*, s.plan
	from public.users u
		join public.subscriptions s on s.user_id = u.id
	order by u.id
	limit $1
	offset $2
	`
	var rows []dao.User
	if err := pgxscan.Select(ctx, s.db.Pool(ctx), &rows, query, limit, offset); err != nil {
		return nil, err
	}
	result := make([]*entity.User, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ConvertTo())
	}
	return result, nil
}
func (s *Storage) Save(ctx context.Context, objects ...*entity.User) error {
	if len(objects) == 0 {
		return nil
	}
	IDs := make([]uuid.UUID, 0, len(objects))
	Name := make([]string, 0, len(objects))
	Email := make([]string, 0, len(objects))
	Version := make([]int, 0, len(objects))
	OldVersion := make([]int, 0, len(objects))
	UpdatedAt := make([]time.Time, 0, len(objects))
	Plan := make([]string, 0, len(objects))
	for _, object := range objects {
		IDs = append(IDs, object.ID())
		Name = append(Name, object.Name())
		Email = append(Email, object.Email())
		Version = append(Version, object.Version())
		OldVersion = append(OldVersion, object.OldVersion())
		UpdatedAt = append(UpdatedAt, object.UpdatedAt())
		Plan = append(Plan, string(object.Plan()))
	}
	const query = `
	update public.users as target
	set
		name = v.name,
		email = v.email,
		version = v.version,
		updated_at = v.updated_at
	from unnest(
		$1::uuid[],
		$2::text[],
		$3::text[],
		$4::bigint[],
		$5::bigint[],
		$6::timestamptz[]
	) as v(id, name, email, version, old_version, updated_at)
	where target.id = v.id
		and target.version = v.old_version
	`
	tag, err := s.db.Pool(ctx).Exec(ctx, query, IDs, Name, Email, Version, OldVersion, UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != int64(len(objects)) {
		return app.NewConflictErr("Данные изменились; перечитайте запись")
	}
	const subscriptions = `
	update public.subscriptions as target
	set
		plan = v.plan,
		updated_at = v.updated_at
	from unnest(
		$1::uuid[],
		$2::text[],
		$3::timestamptz[]
	) as v(user_id, plan, updated_at)
	where target.user_id = v.user_id
	`
	tag, err = s.db.Pool(ctx).Exec(ctx, subscriptions, IDs, Plan, UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != int64(len(objects)) {
		return app.NewInternalError("Не найдена подписка пользователя", nil)
	}
	return nil
}
