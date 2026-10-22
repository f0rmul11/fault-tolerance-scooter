package scooter

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"

	"scooter/internal/domain/entity"
	"scooter/internal/infrastructure/storage/scooter/dao"
	app "scooter/internal/pkg/scootererror"
	"scooter/internal/pkg/transaction/wrapper"
)

type Storage struct {
	db wrapper.Database
}

func New(db wrapper.Database) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Get(ctx context.Context, id uuid.UUID) (*entity.Scooter, error) {
	const query = `
	select *
	from public.scooters
	where id = $1
	`
	var row dao.Scooter
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

func (s *Storage) List(ctx context.Context, limit, offset int) ([]*entity.Scooter, error) {
	const query = `
	select *
	from public.scooters
	order by id
	limit $1
	offset $2
	`
	var rows []dao.Scooter
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
	result := make([]*entity.Scooter, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ConvertTo())
	}
	return result, nil
}

// Save сохраняет изменяемые поля с атомарной проверкой обеих версий.
func (s *Storage) Save(ctx context.Context, objects ...*entity.Scooter) error {
	if len(objects) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(objects))
	active := make([]bool, 0, len(objects))
	owners := make([]string, 0, len(objects))
	versions := make([]int, 0, len(objects))
	oldVersions := make([]int, 0, len(objects))
	expectedVersions := make([]int, 0, len(objects))
	updatedAt := make([]time.Time, 0, len(objects))
	for _, object := range objects {
		ids = append(ids, object.ID())
		active = append(active, object.IsActive())
		owners = append(owners, object.Owner())
		versions = append(versions, object.Version())
		oldVersions = append(oldVersions, object.OldVersion())
		expectedVersions = append(expectedVersions, object.ExpectedVersion())
		updatedAt = append(updatedAt, object.UpdatedAt())
	}
	const query = `
	update public.scooters as target
	set
		is_active = v.is_active,
		owner = v.owner,
		version = v.version,
		updated_at = v.updated_at
	from unnest(
		$1::uuid[],
		$2::boolean[],
		$3::text[],
		$4::bigint[],
		$5::bigint[],
		$6::bigint[],
		$7::timestamptz[]
	) as v(id, is_active, owner, version, old_version, expected_version, updated_at)
	where target.id = v.id
		and target.version = v.old_version
		and target.version = v.expected_version
	`
	tag, err := s.db.Pool(ctx).Exec(
		ctx,
		query,
		ids,
		active,
		owners,
		versions,
		oldVersions,
		expectedVersions,
		updatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != int64(len(objects)) {
		return app.NewConflictErr("Данные изменились; перечитайте запись")
	}
	return nil
}
