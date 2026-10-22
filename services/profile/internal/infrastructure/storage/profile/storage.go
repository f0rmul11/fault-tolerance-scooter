package profile

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"

	"profile/internal/domain/entity"
	"profile/internal/infrastructure/storage/profile/dao"
	app "profile/internal/pkg/profileerror"
	"profile/internal/pkg/transaction/wrapper"
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
	if err := pgxscan.Get(
		ctx,
		s.db.Pool(ctx),
		&row,
		query,
		id,
	); err != nil {
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
	if err := pgxscan.Select(
		ctx,
		s.db.Pool(ctx),
		&rows,
		query,
		limit,
		offset,
	); err != nil {
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
	ids := make([]uuid.UUID, 0, len(objects))
	names := make([]string, 0, len(objects))
	emails := make([]string, 0, len(objects))
	versions := make([]int, 0, len(objects))
	oldVersions := make([]int, 0, len(objects))
	expectedVersions := make([]int, 0, len(objects))
	updatedAt := make([]time.Time, 0, len(objects))
	for _, object := range objects {
		ids = append(ids, object.ID())
		names = append(names, object.Name())
		emails = append(emails, object.Email())
		versions = append(versions, object.Version())
		oldVersions = append(oldVersions, object.OldVersion())
		expectedVersions = append(expectedVersions, object.ExpectedVersion())
		updatedAt = append(updatedAt, object.UpdatedAt())
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
		$6::timestamptz[],
		$7::bigint[]
	) as v(id, name, email, version, old_version, updated_at, expected_version)
	where target.id = v.id
		and target.version = v.old_version
		and target.version = v.expected_version
	`
	tag, err := s.db.Pool(ctx).Exec(
		ctx,
		query,
		ids,
		names,
		emails,
		versions,
		oldVersions,
		updatedAt,
		expectedVersions,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != int64(len(objects)) {
		return app.NewConflictErr("Данные изменились; перечитайте запись")
	}
	return nil
}

// SaveSubscriptions сохраняет подписки после успешного CAS профилей в той же транзакции.
func (s *Storage) SaveSubscriptions(ctx context.Context, objects ...*entity.User) error {
	if len(objects) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(objects))
	plans := make([]string, 0, len(objects))
	updatedAt := make([]time.Time, 0, len(objects))
	for _, object := range objects {
		ids = append(ids, object.ID())
		plans = append(plans, string(object.Plan()))
		updatedAt = append(updatedAt, object.UpdatedAt())
	}
	const query = `
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
	tag, err := s.db.Pool(ctx).Exec(
		ctx,
		query,
		ids,
		plans,
		updatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != int64(len(objects)) {
		return app.NewInternalError("Не найдена подписка пользователя", nil)
	}
	return nil
}
