package scooter

import (
	"context"
	"errors"
	"scooter/internal/domain/entity"
	"scooter/internal/infrastructure/storage/scooter/dao"
	app "scooter/internal/pkg/scootererror"
	"scooter/internal/pkg/transaction/wrapper"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	if err := pgxscan.Get(ctx, s.db.Pool(ctx), &row, query, id); err != nil {
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
	if err := pgxscan.Select(ctx, s.db.Pool(ctx), &rows, query, limit, offset); err != nil {
		return nil, err
	}
	result := make([]*entity.Scooter, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ConvertTo())
	}
	return result, nil
}
func (s *Storage) Save(ctx context.Context, objects ...*entity.Scooter) error {
	if len(objects) == 0 {
		return nil
	}
	IDs := make([]uuid.UUID, 0, len(objects))
	Number := make([]string, 0, len(objects))
	Brand := make([]string, 0, len(objects))
	ChargePercent := make([]int, 0, len(objects))
	RiddenSeconds := make([]int64, 0, len(objects))
	Version := make([]int, 0, len(objects))
	OldVersion := make([]int, 0, len(objects))
	UpdatedAt := make([]time.Time, 0, len(objects))
	for _, object := range objects {
		IDs = append(IDs, object.ID())
		Number = append(Number, object.Number())
		Brand = append(Brand, object.Brand())
		ChargePercent = append(ChargePercent, object.ChargePercent())
		RiddenSeconds = append(RiddenSeconds, object.RiddenSeconds())
		Version = append(Version, object.Version())
		OldVersion = append(OldVersion, object.OldVersion())
		UpdatedAt = append(UpdatedAt, object.UpdatedAt())
	}
	const query = `
	update public.scooters as target
	set
		number = v.number,
		brand = v.brand,
		charge_percent = v.charge_percent,
		ridden_seconds = v.ridden_seconds,
		version = v.version,
		updated_at = v.updated_at
	from unnest(
		$1::uuid[],
		$2::text[],
		$3::text[],
		$4::bigint[],
		$5::bigint[],
		$6::bigint[],
		$7::bigint[],
		$8::timestamptz[]
	) as v(id, number, brand, charge_percent, ridden_seconds, version, old_version, updated_at)
	where target.id = v.id
		and target.version = v.old_version
	`
	tag, err := s.db.Pool(ctx).Exec(ctx, query, IDs, Number, Brand, ChargePercent, RiddenSeconds, Version, OldVersion, UpdatedAt)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" && pgError.ConstraintName == "scooters_number_key" {
			return app.NewConflictErr("Такой номер самоката уже существует")
		}
		return err
	}
	if tag.RowsAffected() != int64(len(objects)) {
		return app.NewConflictErr("Данные изменились; перечитайте запись")
	}
	return nil
}
