package store

import (
	"database/sql"
	"fmt"
	"time"
)

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := formatTime(*t)
	return &s
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing stored time %q: %w", s, err)
	}
	return t, nil
}

func ptr[T any](n sql.Null[T]) *T {
	if !n.Valid {
		return nil
	}
	return &n.V
}
