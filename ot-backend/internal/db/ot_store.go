package db

import (
	"context"
	"fmt"
	"strings"

	"ot-uat/internal/engine"

	"github.com/jackc/pgx/v5"
)

type Staff struct {
	StaffID     string `json:"staffid"`
	NameEng     string `json:"nameeng"`
	NameChi     string `json:"namechi"`
	DisplayName string `json:"displayname"`
	DomainName  string `json:"domainname"`
}

func (s *Store) UpsertStaff(ctx context.Context, staffNo, staffCode string) (Staff, error) {
	updateTag, err := s.pool.Exec(ctx, `
		UPDATE staffinfo.staffinfo
		SET domainname = $2,
		    displayname = $1,
		    nameeng = $1,
		    namechi = $1
		WHERE staffid = $1
	`, staffNo, staffCode)
	if err != nil {
		return Staff{}, err
	}
	if updateTag.RowsAffected() == 0 {
		if _, err := s.pool.Exec(ctx, `
			INSERT INTO staffinfo.staffinfo (staffid, nameeng, namechi, displayname, domainname)
			VALUES ($1, $1, $1, $1, $2)
		`, staffNo, staffCode); err != nil {
			return Staff{}, err
		}
	}

	var out Staff
	if err := s.pool.QueryRow(ctx, `
		SELECT staffid, nameeng, namechi, displayname, domainname
		FROM staffinfo.staffinfo
		WHERE staffid = $1
	`, staffNo).Scan(&out.StaffID, &out.NameEng, &out.NameChi, &out.DisplayName, &out.DomainName); err != nil {
		return Staff{}, err
	}
	return out, nil
}

type EntryInput struct {
	Type      string  `json:"type"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	InputBy   *string `json:"inputBy"`
}

type SavedEntry struct {
	ID        int64   `json:"id"`
	OTID      int64   `json:"otid"`
	OTStaffID string  `json:"otstaffid"`
	Date      string  `json:"date"`
	Period    string  `json:"period"`
	Type      string  `json:"type"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	InputBy   *string `json:"inputBy"`
}

type MonthlyTotal struct {
	Year       int `json:"year"`
	Month      int `json:"month"`
	TotalHrs20 int `json:"totalhrs20"`
	TotalHrs15 int `json:"totalhrs15"`
}

func (s *Store) ListStaff(ctx context.Context) ([]Staff, error) {
	rows, err := s.pool.Query(ctx, `SELECT staffid, nameeng, namechi, displayname, domainname FROM staffinfo.staffinfo ORDER BY displayname, staffid`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Staff{}
	for rows.Next() {
		var r Staff
		if err := rows.Scan(&r.StaffID, &r.NameEng, &r.NameChi, &r.DisplayName, &r.DomainName); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) SavePeriodEntries(ctx context.Context, otstaffid, date, period string, entries []EntryInput) ([]SavedEntry, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	periodID, err := upsertOTPeriod(ctx, tx, otstaffid, date, period)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM otdriverstd.otdetails WHERE otid = $1`, periodID); err != nil {
		return nil, err
	}
	for _, e := range entries {
		if _, err := tx.Exec(ctx, `INSERT INTO otdriverstd.otdetails (otid, type, starttime, endtime, inputby) VALUES ($1, $2, $3::time, $4::time, $5)`, periodID, e.Type, e.StartTime, e.EndTime, nullableTrim(e.InputBy)); err != nil {
			return nil, err
		}
	}
	if err := s.rebuildPeriodResultTx(ctx, tx, periodID, otstaffid, date, period); err != nil {
		return nil, err
	}

	saved, err := getEntriesByFilters(ctx, tx, otstaffid, date, period)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return saved, nil
}

func (s *Store) GetEntries(ctx context.Context, otstaffid, date, period string) ([]SavedEntry, error) {
	return getEntriesByFilters(ctx, s.pool, otstaffid, date, period)
}

func getEntriesByFilters(ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, otstaffid, date, period string) ([]SavedEntry, error) {
	rows, err := q.Query(ctx, `
		SELECT d.id, d.otid, p.otstaffid, to_char(p.date, 'YYYY-MM-DD'), p.period, d.type,
		       to_char(d.starttime, 'HH24:MI'), to_char(d.endtime, 'HH24:MI'), d.inputby
		FROM otdriverstd.otdetails d
		JOIN otdriverstd.otperiod p ON p.id = d.otid
		WHERE ($1 = '' OR p.otstaffid = $1)
		  AND ($2 = '' OR p.date = $2::date)
		  AND ($3 = '' OR p.period = $3)
		ORDER BY p.date, p.period, d.id
	`, otstaffid, date, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []SavedEntry{}
	for rows.Next() {
		var r SavedEntry
		if err := rows.Scan(&r.ID, &r.OTID, &r.OTStaffID, &r.Date, &r.Period, &r.Type, &r.StartTime, &r.EndTime, &r.InputBy); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) GetMonthlyTotals(ctx context.Context, year int, month int) ([]MonthlyTotal, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT EXTRACT(YEAR FROM date_label)::int AS y,
		       EXTRACT(MONTH FROM date_label)::int AS m,
		       COALESCE(SUM(totalhrs20),0)::int,
		       COALESCE(SUM(totalhrs15),0)::int
		FROM otdriverstd.periodresult
		WHERE ($1 = 0 OR EXTRACT(YEAR FROM date_label)::int = $1)
		  AND ($2 = 0 OR EXTRACT(MONTH FROM date_label)::int = $2)
		GROUP BY y, m
		ORDER BY y, m
	`, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []MonthlyTotal{}
	for rows.Next() {
		var r MonthlyTotal
		if err := rows.Scan(&r.Year, &r.Month, &r.TotalHrs20, &r.TotalHrs15); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) rebuildPeriodResultTx(ctx context.Context, tx pgx.Tx, periodID int64, otstaffid, date, period string) error {
	rows, err := tx.Query(ctx, `SELECT id, type, to_char(starttime, 'HH24:MI'), to_char(endtime, 'HH24:MI') FROM otdriverstd.otdetails WHERE otid = $1 ORDER BY id`, periodID)
	if err != nil {
		return err
	}
	defer rows.Close()

	input := engine.CalculateInput{}
	proc20Parts := []string{}
	proc15Parts := []string{}
	for rows.Next() {
		var id int64
		var t, start, end string
		if err := rows.Scan(&id, &t, &start, &end); err != nil {
			return err
		}
		if t == "01" {
			input.BreakEntries = append(input.BreakEntries, engine.BreakEntry{ID: fmt.Sprintf("B%d", id), EmployeeID: engine.EmployeeA, Date: date, Period: periodToEngine(period), StartTime: start, EndTime: end})
			proc15Parts = append(proc15Parts, start+"-"+end)
		} else {
			input.OTEntries = append(input.OTEntries, engine.OTEntry{ID: fmt.Sprintf("O%d", id), EmployeeID: engine.EmployeeA, Date: date, Period: periodToEngine(period), StartTime: start, EndTime: end})
			proc20Parts = append(proc20Parts, start+"-"+end)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	out, err := engine.NewCalculator().Calculate(input)
	if err != nil {
		return err
	}
	daily := engine.DailySummary{DateLabel: date}
	for _, d := range out.DailySummary[engine.EmployeeA] {
		daily = d
		break
	}

	id := makePeriodResultID(date, period)
	process20 := "2.0 process: " + strings.Join(proc20Parts, " + ")
	process15 := "1.5 process: " + strings.Join(proc15Parts, " + ")
	updateTag, err := tx.Exec(ctx, `
		UPDATE otdriverstd.periodresult
		SET otstaffid = $2, date_label = $3::date, process20txt = $4, process15txt = $5,
		    hours20 = $6, hours15 = $7, mins20 = $8, mins15 = $9, totalhrs20 = $10, totalhrs15 = $11, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, id, otstaffid, date, process20, process15,
		daily.Rate20RoundedHours, daily.Rate15RoundedHours, daily.Rate20Minutes, daily.Rate15Minutes,
		daily.Rate20RoundedHours, daily.Rate15RoundedHours)
	if err != nil {
		return err
	}
	if updateTag.RowsAffected() == 0 {
		_, err = tx.Exec(ctx, `
			INSERT INTO otdriverstd.periodresult
			(id, otstaffid, date_label, process20txt, process15txt, hours20, hours15, mins20, mins15, totalhrs20, totalhrs15)
			VALUES ($1, $2, $3::date, $4, $5, $6, $7, $8, $9, $10, $11)
		`, id, otstaffid, date, process20, process15,
			daily.Rate20RoundedHours, daily.Rate15RoundedHours, daily.Rate20Minutes, daily.Rate15Minutes,
			daily.Rate20RoundedHours, daily.Rate15RoundedHours)
		if err != nil {
			return err
		}
	}
	return nil
}

func upsertOTPeriod(ctx context.Context, tx pgx.Tx, otstaffid, date, period string) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `SELECT id FROM otdriverstd.otperiod WHERE otstaffid = $1 AND date = $2::date AND period = $3`, otstaffid, date, period).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != pgx.ErrNoRows {
		return 0, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO otdriverstd.otperiod (date, otstaffid, period, remarks) VALUES ($1::date, $2, $3, '')`, date, otstaffid, period); err != nil {
		return 0, err
	}
	if err := tx.QueryRow(ctx, `SELECT id FROM otdriverstd.otperiod WHERE otstaffid = $1 AND date = $2::date AND period = $3`, otstaffid, date, period).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func nullableTrim(v *string) any {
	if v == nil {
		return nil
	}
	t := strings.TrimSpace(*v)
	if t == "" {
		return nil
	}
	return t
}

func makePeriodResultID(date, period string) string {
	return strings.ReplaceAll(date, "-", "") + period
}

func periodToEngine(period string) string {
	if period == "00" {
		return "AM"
	}
	return "PM"
}
