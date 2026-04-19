package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ot-uat/internal/db"
	"ot-uat/internal/engine"
)

type CalculateService struct {
	calculator *engine.Calculator
	store      *db.Store
}

func NewCalculateService(calculator *engine.Calculator, store *db.Store) *CalculateService {
	return &CalculateService{calculator: calculator, store: store}
}

func (s *CalculateService) Calculate(input engine.CalculateInput) (engine.CalculateOutput, error) {
	normalized, err := normalizeInput(input)
	if err != nil {
		return engine.CalculateOutput{}, err
	}
	out, err := s.calculator.Calculate(normalized)
	if err != nil {
		return engine.CalculateOutput{}, err
	}
	if s.store != nil {
		if err := s.store.SaveCalculation(normalized, out); err != nil {
			return engine.CalculateOutput{}, err
		}
	}
	_ = writeSessionCache(out)
	return out, nil
}

func normalizeInput(input engine.CalculateInput) (engine.CalculateInput, error) {
	out := engine.CalculateInput{
		OTEntries:    make([]engine.OTEntry, 0, len(input.OTEntries)),
		BreakEntries: make([]engine.BreakEntry, 0, len(input.BreakEntries)),
	}

	for _, e := range input.OTEntries {
		date, period, start, end, err := validateAndNormalizeCommon(e.EmployeeID, e.Date, e.Period, e.StartTime, e.EndTime)
		if err != nil {
			return engine.CalculateInput{}, fmt.Errorf("ot entry %s: %w", e.ID, err)
		}
		e.Date, e.Period, e.StartTime, e.EndTime = date, period, start, end
		out.OTEntries = append(out.OTEntries, e)
	}
	for _, e := range input.BreakEntries {
		date, period, start, end, err := validateAndNormalizeCommon(e.EmployeeID, e.Date, e.Period, e.StartTime, e.EndTime)
		if err != nil {
			return engine.CalculateInput{}, fmt.Errorf("break entry %s: %w", e.ID, err)
		}
		e.Date, e.Period, e.StartTime, e.EndTime = date, period, start, end
		out.BreakEntries = append(out.BreakEntries, e)
	}
	return out, nil
}

func validateAndNormalizeCommon(employeeID engine.EmployeeID, date, period, start, end string) (string, string, string, string, error) {
	if employeeID != engine.EmployeeA && employeeID != engine.EmployeeB {
		return "", "", "", "", errors.New("employeeId must be A or B")
	}
	dateNorm, err := parseDate(date)
	if err != nil {
		return "", "", "", "", errors.New("date must be YYYY-MM-DD or MM/DD/YYYY")
	}
	periodNorm, err := parsePeriod(period)
	if err != nil {
		return "", "", "", "", errors.New("period must be AM or PM")
	}
	startNorm, err := parse24Hour(start)
	if err != nil {
		return "", "", "", "", errors.New("startTime must be 24-hour HH:MM")
	}
	endNorm, err := parse24Hour(end)
	if err != nil {
		return "", "", "", "", errors.New("endTime must be 24-hour HH:MM")
	}
	return dateNorm, periodNorm, startNorm, endNorm, nil
}

func parseDate(raw string) (string, error) {
	for _, layout := range []string{"2006-01-02", "01/02/2006", "1/2/2006"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", errors.New("invalid date")
}

func parse24Hour(raw string) (string, error) {
	t, err := time.Parse("15:04", raw)
	if err != nil {
		return "", err
	}
	return t.Format("15:04"), nil
}

func parsePeriod(raw string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "AM":
		return "AM", nil
	case "PM":
		return "PM", nil
	default:
		return "", errors.New("invalid period")
	}
}

type cacheRecord struct {
	DateLabel          string  `json:"dateLabel"`
	Rate15RoundedHours int     `json:"rate15RoundedHours"`
	Rate20RoundedHours int     `json:"rate20RoundedHours"`
	TotalWeighted      float64 `json:"totalWeighted"`
}

func writeSessionCache(out engine.CalculateOutput) error {
	cache := map[string]map[string]cacheRecord{}
	for emp, byKey := range out.DailySummary {
		empKey := string(emp)
		for sessionKey, daily := range byKey {
			if _, ok := cache[sessionKey]; !ok {
				cache[sessionKey] = map[string]cacheRecord{}
			}
			cache[sessionKey][empKey] = cacheRecord{
				DateLabel:          daily.DateLabel,
				Rate15RoundedHours: daily.Rate15RoundedHours,
				Rate20RoundedHours: daily.Rate20RoundedHours,
				TotalWeighted:      daily.TotalWeighted,
			}
		}
	}

	baseDir := filepath.Join(os.TempDir(), "ot-uat")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return err
	}
	file := filepath.Join(baseDir, "session_summary_cache.json")
	b, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, b, 0o644)
}
