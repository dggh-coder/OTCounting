package service

import (
	"errors"
	"fmt"
	"time"

	"ot-uat/internal/engine"
)

type CalculateService struct {
	calculator *engine.Calculator
}

func NewCalculateService(calculator *engine.Calculator) *CalculateService {
	return &CalculateService{calculator: calculator}
}

func (s *CalculateService) Calculate(input engine.CalculateInput) (engine.CalculateOutput, error) {
	normalized, err := normalizeInput(input)
	if err != nil {
		return engine.CalculateOutput{}, err
	}
	return s.calculator.Calculate(normalized)
}

func normalizeInput(input engine.CalculateInput) (engine.CalculateInput, error) {
	out := engine.CalculateInput{
		OTEntries:    make([]engine.OTEntry, 0, len(input.OTEntries)),
		BreakEntries: make([]engine.BreakEntry, 0, len(input.BreakEntries)),
	}

	for _, e := range input.OTEntries {
		date, start, end, err := validateAndNormalizeCommon(e.EmployeeID, e.Date, e.StartTime, e.EndTime)
		if err != nil {
			return engine.CalculateInput{}, fmt.Errorf("ot entry %s: %w", e.ID, err)
		}
		e.Date, e.StartTime, e.EndTime = date, start, end
		out.OTEntries = append(out.OTEntries, e)
	}
	for _, e := range input.BreakEntries {
		date, start, end, err := validateAndNormalizeCommon(e.EmployeeID, e.Date, e.StartTime, e.EndTime)
		if err != nil {
			return engine.CalculateInput{}, fmt.Errorf("break entry %s: %w", e.ID, err)
		}
		e.Date, e.StartTime, e.EndTime = date, start, end
		out.BreakEntries = append(out.BreakEntries, e)
	}
	return out, nil
}

func validateAndNormalizeCommon(employeeID engine.EmployeeID, date, start, end string) (string, string, string, error) {
	if employeeID != engine.EmployeeA && employeeID != engine.EmployeeB {
		return "", "", "", errors.New("employeeId must be A or B")
	}
	dateNorm, err := parseDate(date)
	if err != nil {
		return "", "", "", errors.New("date must be YYYY-MM-DD or MM/DD/YYYY")
	}
	startNorm, err := parse24Hour(start)
	if err != nil {
		return "", "", "", errors.New("startTime must be 24-hour HH:MM")
	}
	endNorm, err := parse24Hour(end)
	if err != nil {
		return "", "", "", errors.New("endTime must be 24-hour HH:MM")
	}
	return dateNorm, startNorm, endNorm, nil
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
