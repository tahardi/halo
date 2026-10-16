package store

import (
	"context"
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/tahardi/halo/backend/internal/model"
)

func (s *Store) PutMedication(ctx context.Context, m model.Medication) (model.Medication, error) {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO medications (id, name, usual_amount, unit, kind, archived)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			name = excluded.name,
			usual_amount = excluded.usual_amount,
			unit = excluded.unit,
			kind = excluded.kind,
			archived = excluded.archived`,
		m.ID, m.Name, m.UsualAmount, m.Unit, string(m.Kind), m.Archived,
	)
	if err != nil {
		if isUnique(err) {
			return model.Medication{}, fmt.Errorf("%w: medication named %q", ErrDuplicate, m.Name)
		}
		return model.Medication{}, fmt.Errorf("saving medication: %w", err)
	}
	return m, nil
}

func (s *Store) ListMedications(ctx context.Context) ([]model.Medication, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, usual_amount, unit, kind, archived
		FROM medications
		ORDER BY name COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("listing medications: %w", err)
	}
	defer rows.Close()
	meds := []model.Medication{}
	for rows.Next() {
		var m model.Medication
		if err := rows.Scan(&m.ID, &m.Name, &m.UsualAmount, &m.Unit, &m.Kind, &m.Archived); err != nil {
			return nil, fmt.Errorf("scanning medication: %w", err)
		}
		meds = append(meds, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading medications: %w", err)
	}
	return meds, nil
}

func isUnique(err error) bool {
	sqliteErr, ok := errors.AsType[*sqlite.Error](err)
	return ok && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}
