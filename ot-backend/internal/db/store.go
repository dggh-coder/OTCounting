package db

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"ot-uat/internal/engine"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(dsn string) (*Store, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) InitSchema() error {
	_, err := s.pool.Exec(context.Background(), schemaSQL)
	return err
}

func (s *Store) SaveCalculation(input engine.CalculateInput, out engine.CalculateOutput) error {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sessionSeen := map[string]struct{}{}
	for _, e := range appendOTAndBreak(input) {
		period := strings.ToUpper(strings.TrimSpace(e.Period))
		sessionID := makeSessionID(e.Date, period)
		if _, ok := sessionSeen[sessionID]; !ok {
			_, err = tx.Exec(ctx, `
				INSERT INTO ot_uat.work_session (session_id, session_date, period, status, created_by)
				SELECT $1, $2::date, $3, 'OPEN', 'ot-backend'
				WHERE NOT EXISTS (
					SELECT 1 FROM ot_uat.work_session WHERE session_id = $1
				)
			`, sessionID, e.Date, period)
			if err != nil {
				return err
			}
			sessionSeen[sessionID] = struct{}{}
		}
		updateTag, err := tx.Exec(ctx, `
			UPDATE ot_uat.time_entry
			SET
				session_id = $2,
				employee_id = $3,
				entry_type = $4,
				start_time = $5::time,
				end_time = $6::time,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`, e.ID, sessionID, e.EmployeeID, e.EntryType, e.StartTime, e.EndTime)
		if err != nil {
			return err
		}

		if updateTag.RowsAffected() == 0 {
			_, err = tx.Exec(ctx, `
				INSERT INTO ot_uat.time_entry (id, session_id, employee_id, entry_type, start_time, end_time)
				VALUES ($1, $2, $3, $4, $5::time, $6::time)
			`, e.ID, sessionID, e.EmployeeID, e.EntryType, e.StartTime, e.EndTime)
		}
		if err != nil {
			return err
		}
	}

	for emp, bySession := range out.DailySummary {
		for sessionID, daily := range bySession {
			updateTag, err := tx.Exec(ctx, `
				UPDATE ot_uat.session_result
				SET
					date_label = $3,
					rate20_rounded_hours = $4,
					rate15_rounded_hours = $5,
					rate20_minutes = $6,
					rate15_minutes = $7,
					calculated_at = $8
				WHERE session_id = $1 AND employee_id = $2
			`, sessionID, emp, daily.DateLabel, daily.Rate20RoundedHours, daily.Rate15RoundedHours,
				daily.Rate20Minutes, daily.Rate15Minutes, time.Now())
			if err != nil {
				return err
			}

			if updateTag.RowsAffected() == 0 {
				_, err = tx.Exec(ctx, `
					INSERT INTO ot_uat.session_result (
						session_id, employee_id, date_label, rate20_rounded_hours, rate15_rounded_hours,
						rate20_minutes, rate15_minutes, calculated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				`, sessionID, emp, daily.DateLabel, daily.Rate20RoundedHours, daily.Rate15RoundedHours,
					daily.Rate20Minutes, daily.Rate15Minutes, time.Now())
			}
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

type dbEntry struct {
	ID         string
	EmployeeID engine.EmployeeID
	Date       string
	Period     string
	StartTime  string
	EndTime    string
	EntryType  string
}

func appendOTAndBreak(input engine.CalculateInput) []dbEntry {
	all := make([]dbEntry, 0, len(input.OTEntries)+len(input.BreakEntries))
	for _, e := range input.OTEntries {
		all = append(all, dbEntry{ID: e.ID, EmployeeID: e.EmployeeID, Date: e.Date, Period: e.Period, StartTime: e.StartTime, EndTime: e.EndTime, EntryType: "OT"})
	}
	for _, e := range input.BreakEntries {
		all = append(all, dbEntry{ID: e.ID, EmployeeID: e.EmployeeID, Date: e.Date, Period: e.Period, StartTime: e.StartTime, EndTime: e.EndTime, EntryType: "BREAK"})
	}
	return all
}

func makeSessionID(date, period string) string {
	d := strings.ReplaceAll(date, "-", "")
	if period == "PM" {
		return d + "02"
	}
	return d + "01"
}
