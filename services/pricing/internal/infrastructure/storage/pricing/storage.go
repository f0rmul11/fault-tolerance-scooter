package pricing

import (
	"context"
	"errors"
	"pricing/internal/domain/entity"
	"pricing/internal/infrastructure/storage/pricing/dao"
	app "pricing/internal/pkg/pricingerror"
	"pricing/internal/pkg/transaction/wrapper"
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
func (s *Storage) Get(ctx context.Context, id uuid.UUID) (*entity.Trip, error) {
	const query = `
	select *
	from public.trips
	where rental_id = $1
	`
	var row dao.Trip
	if err := pgxscan.Get(ctx, s.db.Pool(ctx), &row, query, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app.NewNotFoundErr("Запись не найдена", err)
		}
		return nil, err
	}
	return row.ConvertTo(), nil
}
func (s *Storage) Create(ctx context.Context, object *entity.Trip) error {
	const query = `
	insert into public.trips (
		rental_id,
		user_id,
		scooter_id,
		plan,
		rate_minor,
		advance_minor,
		charged_minor,
		total_minor,
		adjustment_minor,
		billed_minutes,
		status,
		rental_version,
		reserved_at,
		started_at,
		finished_at,
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
		$13,
		$14,
		$15,
		$16
	)
	on conflict do nothing
	`
	tag, err := s.db.Pool(ctx).Exec(ctx, query,
		object.RentalID(),
		object.UserID(),
		object.ScooterID(),
		object.Plan(),
		object.RateMinor(),
		object.AdvanceMinor(),
		object.ChargedMinor(),
		object.TotalMinor(),
		object.AdjustmentMinor(),
		object.BilledMinutes(),
		object.Status(),
		object.RentalVersion(),
		object.ReservedAt(),
		object.StartedAt(),
		object.FinishedAt(),
		object.Version(),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return app.NewConflictErr("Расчёт поездки уже создан; повторите обработку")
	}
	return nil
}
func (s *Storage) Save(ctx context.Context, objects ...*entity.Trip) error {
	if len(objects) == 0 {
		return nil
	}
	RentalID := make([]uuid.UUID, 0, len(objects))
	UserID := make([]uuid.UUID, 0, len(objects))
	ScooterID := make([]uuid.UUID, 0, len(objects))
	Plan := make([]string, 0, len(objects))
	RateMinor := make([]int64, 0, len(objects))
	AdvanceMinor := make([]int64, 0, len(objects))
	ChargedMinor := make([]int64, 0, len(objects))
	TotalMinor := make([]int64, 0, len(objects))
	AdjustmentMinor := make([]int64, 0, len(objects))
	BilledMinutes := make([]int64, 0, len(objects))
	Status := make([]string, 0, len(objects))
	RentalVersion := make([]int, 0, len(objects))
	ReservedAt := make([]time.Time, 0, len(objects))
	StartedAt := make([]time.Time, 0, len(objects))
	FinishedAt := make([]time.Time, 0, len(objects))
	Version := make([]int, 0, len(objects))
	oldVersions := make([]int, 0, len(objects))
	for _, object := range objects {
		RentalID = append(RentalID, object.RentalID())
		UserID = append(UserID, object.UserID())
		ScooterID = append(ScooterID, object.ScooterID())
		Plan = append(Plan, object.Plan())
		RateMinor = append(RateMinor, object.RateMinor())
		AdvanceMinor = append(AdvanceMinor, object.AdvanceMinor())
		ChargedMinor = append(ChargedMinor, object.ChargedMinor())
		TotalMinor = append(TotalMinor, object.TotalMinor())
		AdjustmentMinor = append(AdjustmentMinor, object.AdjustmentMinor())
		BilledMinutes = append(BilledMinutes, object.BilledMinutes())
		Status = append(Status, string(object.Status()))
		RentalVersion = append(RentalVersion, object.RentalVersion())
		ReservedAt = append(ReservedAt, object.ReservedAt())
		StartedAt = append(StartedAt, object.StartedAt())
		FinishedAt = append(FinishedAt, object.FinishedAt())
		Version = append(Version, object.Version())
		oldVersions = append(oldVersions, object.OldVersion())
	}
	const query = `
	update public.trips as target
	set
		user_id = v.user_id,
		scooter_id = v.scooter_id,
		plan = v.plan,
		rate_minor = v.rate_minor,
		advance_minor = v.advance_minor,
		charged_minor = v.charged_minor,
		total_minor = v.total_minor,
		adjustment_minor = v.adjustment_minor,
		billed_minutes = v.billed_minutes,
		status = v.status,
		rental_version = v.rental_version,
		reserved_at = v.reserved_at,
		started_at = v.started_at,
		finished_at = v.finished_at,
		version = v.version
	from unnest(
		$1::uuid[],
		$2::uuid[],
		$3::uuid[],
		$4::text[],
		$5::bigint[],
		$6::bigint[],
		$7::bigint[],
		$8::bigint[],
		$9::bigint[],
		$10::bigint[],
		$11::text[],
		$12::bigint[],
		$13::timestamptz[],
		$14::timestamptz[],
		$15::timestamptz[],
		$16::bigint[],
		$17::bigint[]
	) as v(rental_id, user_id, scooter_id, plan, rate_minor, advance_minor, charged_minor, total_minor, adjustment_minor, billed_minutes, status, rental_version, reserved_at, started_at, finished_at, version, old_version)
	where target.rental_id = v.rental_id
		and target.version = v.old_version
	`
	tag, err := s.db.Pool(ctx).Exec(ctx, query,
		RentalID, UserID, ScooterID, Plan, RateMinor, AdvanceMinor, ChargedMinor, TotalMinor, AdjustmentMinor, BilledMinutes, Status, RentalVersion, ReservedAt, StartedAt, FinishedAt, Version, oldVersions,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != int64(len(objects)) {
		return app.NewConflictErr("Версия записи изменилась; повторите операцию")
	}
	return nil
}
