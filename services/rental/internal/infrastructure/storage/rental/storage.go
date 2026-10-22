package rental

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"

	"rental/internal/domain/entity"
	"rental/internal/infrastructure/storage/rental/dao"
	app "rental/internal/pkg/rentalerror"
	"rental/internal/pkg/transaction/wrapper"
)

type Storage struct {
	db wrapper.Database
}

func New(db wrapper.Database) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Get(ctx context.Context, id uuid.UUID) (*entity.Rental, error) {
	const query = `
	select *
	from public.rentals
	where id = $1
	`
	var row dao.Rental
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

func (s *Storage) FindKey(ctx context.Context, userID, key uuid.UUID) (*entity.Rental, error) {
	const query = `
	select *
	from public.rentals
	where user_id = $1
		and idempotency_key = $2
	`
	var row dao.Rental
	if err := pgxscan.Get(
		ctx,
		s.db.Pool(ctx),
		&row,
		query,
		userID,
		key,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return row.ConvertTo(), nil
}

func (s *Storage) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Rental, error) {
	const query = `
	select *
	from public.rentals
	where user_id = $1
	order by id
	limit $2
	offset $3
	`
	var rows []dao.Rental
	if err := pgxscan.Select(
		ctx,
		s.db.Pool(ctx),
		&rows,
		query,
		userID,
		limit,
		offset,
	); err != nil {
		return nil, err
	}
	result := make([]*entity.Rental, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ConvertTo())
	}
	return result, nil
}

func (s *Storage) Create(ctx context.Context, object *entity.Rental) error {
	const query = `
	insert into public.rentals (
		id,
		user_id,
		scooter_id,
		idempotency_key,
		request_id,
		scooter_number,
		scooter_brand,
		status,
		reserved_at,
		started_at,
		finished_at,
		updated_at,
		version
	) values (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11,
		$12,
		$13
	)
	on conflict do nothing
	`
	tag, err := s.db.Pool(ctx).Exec(
		ctx,
		query,
		object.ID(),
		object.UserID(),
		object.ScooterID(),
		object.IdempotencyKey(),
		object.RequestID(),
		object.ScooterNumber(),
		object.ScooterBrand(),
		object.Status(),
		object.ReservedAt(),
		object.StartedAt(),
		object.FinishedAt(),
		object.UpdatedAt(),
		object.Version(),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return app.NewConflictErr("Самокат, пользователь или ключ уже занят другой арендой")
	}
	return nil
}

func (s *Storage) Save(ctx context.Context, objects ...*entity.Rental) error {
	if len(objects) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(objects))
	userIDs := make([]uuid.UUID, 0, len(objects))
	scooterIDs := make([]uuid.UUID, 0, len(objects))
	idempotencyKeys := make([]uuid.UUID, 0, len(objects))
	requestIDs := make([]uuid.UUID, 0, len(objects))
	scooterNumbers := make([]string, 0, len(objects))
	scooterBrands := make([]string, 0, len(objects))
	statuses := make([]string, 0, len(objects))
	reservedAt := make([]time.Time, 0, len(objects))
	startedAt := make([]time.Time, 0, len(objects))
	finishedAt := make([]time.Time, 0, len(objects))
	updatedAt := make([]time.Time, 0, len(objects))
	versions := make([]int, 0, len(objects))
	oldVersions := make([]int, 0, len(objects))
	for _, object := range objects {
		ids = append(ids, object.ID())
		userIDs = append(userIDs, object.UserID())
		scooterIDs = append(scooterIDs, object.ScooterID())
		idempotencyKeys = append(idempotencyKeys, object.IdempotencyKey())
		requestIDs = append(requestIDs, object.RequestID())
		scooterNumbers = append(scooterNumbers, object.ScooterNumber())
		scooterBrands = append(scooterBrands, object.ScooterBrand())
		statuses = append(statuses, string(object.Status()))
		reservedAt = append(reservedAt, object.ReservedAt())
		startedAt = append(startedAt, object.StartedAt())
		finishedAt = append(finishedAt, object.FinishedAt())
		updatedAt = append(updatedAt, object.UpdatedAt())
		versions = append(versions, object.Version())
		oldVersions = append(oldVersions, object.OldVersion())
	}
	const query = `
	update public.rentals as target
	set
		user_id = v.user_id,
		scooter_id = v.scooter_id,
		idempotency_key = v.idempotency_key,
		request_id = v.request_id,
		scooter_number = v.scooter_number,
		scooter_brand = v.scooter_brand,
		status = v.status,
		reserved_at = v.reserved_at,
		started_at = v.started_at,
		finished_at = v.finished_at,
		updated_at = v.updated_at,
		version = v.version
	from unnest(
		$1::uuid[],
		$2::uuid[],
		$3::uuid[],
		$4::uuid[],
		$5::uuid[],
		$6::text[],
		$7::text[],
		$8::text[],
		$9::timestamptz[],
		$10::timestamptz[],
		$11::timestamptz[],
		$12::timestamptz[],
		$13::bigint[],
		$14::bigint[]
	) as v(id, user_id, scooter_id, idempotency_key, request_id, scooter_number, scooter_brand, status, reserved_at, started_at, finished_at, updated_at, version, old_version)
	where target.id = v.id
		and target.version = v.old_version
	`
	tag, err := s.db.Pool(ctx).Exec(
		ctx,
		query,
		ids,
		userIDs,
		scooterIDs,
		idempotencyKeys,
		requestIDs,
		scooterNumbers,
		scooterBrands,
		statuses,
		reservedAt,
		startedAt,
		finishedAt,
		updatedAt,
		versions,
		oldVersions,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != int64(len(objects)) {
		return app.NewConflictErr("Версия записи изменилась; повторите операцию")
	}
	return nil
}
