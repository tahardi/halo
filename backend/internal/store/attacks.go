package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tahardi/halo/backend/internal/model"
)

const attackColumns = `id, started_at, ended_at, timezone, peak_severity, symptoms, symptoms_other, notes,
	lat, lon, source, created_at, updated_at`

type scanner interface {
	Scan(dest ...any) error
}

func (s *Store) PutAttack(ctx context.Context, a model.Attack) (model.Attack, error) {
	symptoms := a.Symptoms
	if symptoms == nil {
		symptoms = []model.Symptom{}
	}
	symptomsJSON, err := json.Marshal(symptoms)
	if err != nil {
		return model.Attack{}, fmt.Errorf("encoding symptoms: %w", err)
	}
	source := a.Source
	if source == "" {
		source = "app"
	}
	now := formatTime(s.now())

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Attack{}, fmt.Errorf("starting attack save: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO attacks (`+attackColumns+`)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			started_at = excluded.started_at,
			ended_at = excluded.ended_at,
			timezone = excluded.timezone,
			peak_severity = excluded.peak_severity,
			symptoms = excluded.symptoms,
			symptoms_other = excluded.symptoms_other,
			notes = excluded.notes,
			lat = excluded.lat,
			lon = excluded.lon,
			updated_at = excluded.updated_at,
			deleted_at = NULL`,
		a.ID,
		formatTime(a.StartedAt),
		formatTimePtr(a.EndedAt),
		a.Timezone,
		a.PeakSeverity,
		string(symptomsJSON),
		a.SymptomsOther,
		a.Notes,
		a.Lat,
		a.Lon,
		source,
		now,
		now,
	)
	if err != nil {
		return model.Attack{}, fmt.Errorf("saving attack: %w", err)
	}
	if err := putDoses(ctx, tx, a.ID, a.Doses, now); err != nil {
		return model.Attack{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.Attack{}, fmt.Errorf("committing attack: %w", err)
	}
	return s.GetAttack(ctx, a.ID)
}

func putDoses(ctx context.Context, tx *sql.Tx, attackID string, doses []model.Dose, now string) error {
	_, err := tx.ExecContext(
		ctx,
		`UPDATE doses SET deleted_at = ? WHERE attack_id = ? AND deleted_at IS NULL`,
		now,
		attackID,
	)
	if err != nil {
		return fmt.Errorf("clearing doses: %w", err)
	}
	for _, d := range doses {
		var relief *string
		if d.Relief != nil {
			r := string(*d.Relief)
			relief = &r
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO doses (id, attack_id, medication_id, amount, taken_at, relief)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (id) DO UPDATE SET
				medication_id = excluded.medication_id,
				amount = excluded.amount,
				taken_at = excluded.taken_at,
				relief = excluded.relief,
				deleted_at = NULL
			WHERE doses.attack_id = excluded.attack_id`,
			d.ID,
			attackID,
			d.MedicationID,
			d.Amount,
			formatTime(d.TakenAt),
			relief,
		)
		if err != nil {
			return fmt.Errorf("saving dose %s: %w", d.ID, err)
		}
	}
	return nil
}

func (s *Store) GetAttack(ctx context.Context, id string) (model.Attack, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT `+attackColumns+` FROM attacks WHERE id = ? AND deleted_at IS NULL`,
		id,
	)
	a, err := scanAttack(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Attack{}, fmt.Errorf("%w: attack %s", ErrNotFound, id)
	}
	if err != nil {
		return model.Attack{}, err
	}
	a.Doses, err = s.doses(ctx, id)
	if err != nil {
		return model.Attack{}, err
	}
	return a, nil
}

func (s *Store) ListAttacks(ctx context.Context, before time.Time, limit int) ([]model.Attack, error) {
	return s.queryAttacks(ctx, `
		SELECT `+attackColumns+` FROM attacks
		WHERE deleted_at IS NULL AND started_at < ?
		ORDER BY started_at DESC, id DESC
		LIMIT ?`,
		formatTime(before),
		limit,
	)
}

func (s *Store) AllAttacks(ctx context.Context) ([]model.Attack, error) {
	return s.queryAttacks(ctx, `
		SELECT `+attackColumns+` FROM attacks
		WHERE deleted_at IS NULL
		ORDER BY started_at, id`)
}

func (s *Store) DeleteAttack(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(
		ctx,
		`UPDATE attacks SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`,
		formatTime(s.now()),
		id,
	)
	if err != nil {
		return fmt.Errorf("deleting attack: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking deleted attack: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("%w: attack %s", ErrNotFound, id)
	}
	return nil
}

func (s *Store) queryAttacks(ctx context.Context, query string, args ...any) ([]model.Attack, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing attacks: %w", err)
	}
	attacks := []model.Attack{}
	for rows.Next() {
		a, err := scanAttack(rows)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		attacks = append(attacks, a)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("reading attacks: %w", err)
	}
	_ = rows.Close()
	for i := range attacks {
		attacks[i].Doses, err = s.doses(ctx, attacks[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return attacks, nil
}

func scanAttack(row scanner) (model.Attack, error) {
	var (
		a                              model.Attack
		startedAt, symptoms, createdAt string
		updatedAt                      string
		endedAt, symptomsOther, notes  sql.Null[string]
		severity                       sql.Null[int]
		lat, lon                       sql.Null[float64]
	)
	err := row.Scan(
		&a.ID,
		&startedAt,
		&endedAt,
		&a.Timezone,
		&severity,
		&symptoms,
		&symptomsOther,
		&notes,
		&lat,
		&lon,
		&a.Source,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Attack{}, err
	}
	if err != nil {
		return model.Attack{}, fmt.Errorf("scanning attack: %w", err)
	}
	if a.StartedAt, err = parseTime(startedAt); err != nil {
		return model.Attack{}, err
	}
	if endedAt.Valid {
		ended, err := parseTime(endedAt.V)
		if err != nil {
			return model.Attack{}, err
		}
		a.EndedAt = &ended
	}
	if a.CreatedAt, err = parseTime(createdAt); err != nil {
		return model.Attack{}, err
	}
	if a.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return model.Attack{}, err
	}
	if err := json.Unmarshal([]byte(symptoms), &a.Symptoms); err != nil {
		return model.Attack{}, fmt.Errorf("decoding symptoms: %w", err)
	}
	a.PeakSeverity = ptr(severity)
	a.SymptomsOther = ptr(symptomsOther)
	a.Notes = ptr(notes)
	a.Lat = ptr(lat)
	a.Lon = ptr(lon)
	return a, nil
}

func (s *Store) doses(ctx context.Context, attackID string) ([]model.Dose, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, medication_id, amount, taken_at, relief FROM doses
		WHERE attack_id = ? AND deleted_at IS NULL
		ORDER BY taken_at, id`,
		attackID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing doses: %w", err)
	}
	defer rows.Close()
	doses := []model.Dose{}
	for rows.Next() {
		var (
			d       model.Dose
			takenAt string
			relief  sql.Null[string]
		)
		if err := rows.Scan(&d.ID, &d.MedicationID, &d.Amount, &takenAt, &relief); err != nil {
			return nil, fmt.Errorf("scanning dose: %w", err)
		}
		if d.TakenAt, err = parseTime(takenAt); err != nil {
			return nil, err
		}
		if relief.Valid {
			r := model.Relief(relief.V)
			d.Relief = &r
		}
		doses = append(doses, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading doses: %w", err)
	}
	return doses, nil
}
