package service

import (
	"errors"
	"fmt"

	"ot-uat/internal/engine"
)

type CalculateService struct {
	calculator *engine.Calculator
}

func NewCalculateService(calculator *engine.Calculator) *CalculateService {
	return &CalculateService{calculator: calculator}
}

func (s *CalculateService) Calculate(input engine.CalculateInput) (engine.CalculateOutput, error) {
	if err := validateInput(input); err != nil {
		return engine.CalculateOutput{}, err
	}
	return s.calculator.Calculate(input)
}

func validateInput(input engine.CalculateInput) error {
	for _, e := range input.OTEntries {
		if err := validateCommon(e.EmployeeID, e.Date, e.StartTime, e.EndTime); err != nil {
			return fmt.Errorf("ot entry %s: %w", e.ID, err)
		}
	}
	for _, e := range input.BreakEntries {
		if err := validateCommon(e.EmployeeID, e.Date, e.StartTime, e.EndTime); err != nil {
			return fmt.Errorf("break entry %s: %w", e.ID, err)
		}
	}
	return nil
}

func validateCommon(employeeID engine.EmployeeID, date, start, end string) error {
	if employeeID != engine.EmployeeA && employeeID != engine.EmployeeB {
		return errors.New("employeeId must be A or B")
	}
	if len(date) != len("2006-01-02") {
		return errors.New("date must be YYYY-MM-DD")
	}
	if len(start) != len("15:04") || len(end) != len("15:04") {
		return errors.New("time must be HH:MM")
	}
	return nil
}
