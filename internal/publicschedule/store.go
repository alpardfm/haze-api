package publicschedule

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SQLStore struct {
	DB       *sql.DB
	Timezone string
}

func (s SQLStore) ListOccupiedByDate(ctx context.Context, date time.Time) ([]OccupiedRange, error) {
	timezone := s.Timezone
	if timezone == "" {
		timezone = "Asia/Jakarta"
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT
			to_char(start_at AT TIME ZONE $2, 'HH24:MI') AS start_time,
			to_char(end_at AT TIME ZONE $2, 'HH24:MI') AS end_time
		FROM appointments
		WHERE meeting_date = $1
			AND status IN ('scheduled', 'on_going')
		ORDER BY start_at ASC
	`, date, timezone)
	if err != nil {
		return nil, fmt.Errorf("list occupied public schedules: %w", err)
	}
	defer rows.Close()

	var items []OccupiedRange
	for rows.Next() {
		var item OccupiedRange
		if err := rows.Scan(&item.Start, &item.End); err != nil {
			return nil, fmt.Errorf("scan occupied public schedule: %w", err)
		}
		item.Status = "occupied"
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate occupied public schedules: %w", err)
	}

	return items, nil
}
