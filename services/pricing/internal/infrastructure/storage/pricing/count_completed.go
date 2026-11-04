package pricing

import (
	"context"

	"github.com/gofrs/uuid"
)

// CountCompleted читает только собственную историю завершённых поездок.
func (s *Storage) CountCompleted(ctx context.Context, userID uuid.UUID) (int64, error) {
	const query = `
 select count(*)
 from public.trips
 where user_id = $1 and status = 'COMPLETED'
 `
	var count int64
	err := s.db.Pool(ctx).QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}
