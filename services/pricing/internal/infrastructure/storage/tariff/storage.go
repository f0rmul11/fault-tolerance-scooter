package tariff

import (
	"context"
	"errors"
	"pricing/internal/dto/tariff"
	"pricing/internal/infrastructure/storage/tariff/dao"
	app "pricing/internal/pkg/pricingerror"
	"pricing/internal/pkg/transaction/wrapper"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type Storage struct {
	db wrapper.Database
}

func New(db wrapper.Database) *Storage {
	return &Storage{db: db}
}
func (s *Storage) Get(ctx context.Context, plan string) (tariff.Tariff, error) {
	const query = `
	select *
	from public.tariffs
	where plan = $1
	`
	var row dao.Tariff
	if err := pgxscan.Get(ctx, s.db.Pool(ctx), &row, query, plan); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tariff.Tariff{}, app.NewNotFoundErr("Тариф не найден", err)
		}
		return tariff.Tariff{}, err
	}
	return row.ConvertTo(), nil
}
func (s *Storage) List(ctx context.Context) ([]tariff.Tariff, error) {
	const query = `
	select *
	from public.tariffs
	order by plan
	`
	var rows []dao.Tariff
	if err := pgxscan.Select(ctx, s.db.Pool(ctx), &rows, query); err != nil {
		return nil, err
	}
	result := make([]tariff.Tariff, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ConvertTo())
	}
	return result, nil
}
