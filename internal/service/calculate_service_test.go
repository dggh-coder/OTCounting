package service

import (
	"testing"

	"ot-uat/internal/engine"
)

func TestNormalizeInputAcceptsSlashDate(t *testing.T) {
	svc := NewCalculateService(engine.NewCalculator())
	out, err := svc.Calculate(engine.CalculateInput{
		OTEntries: []engine.OTEntry{{
			ID:         "ot-1",
			EmployeeID: engine.EmployeeA,
			Date:       "04/16/2026",
			StartTime:  "20:00",
			EndTime:    "21:00",
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	day, ok := out.DailySummary[engine.EmployeeA]["2026-04-16"]
	if !ok || day.Rate20Minutes != 60 {
		t.Fatalf("unexpected summary: %+v", out.DailySummary[engine.EmployeeA])
	}
}

func TestRejectsNon24HourTime(t *testing.T) {
	svc := NewCalculateService(engine.NewCalculator())
	_, err := svc.Calculate(engine.CalculateInput{
		OTEntries: []engine.OTEntry{{
			ID:         "ot-1",
			EmployeeID: engine.EmployeeA,
			Date:       "2026-04-16",
			StartTime:  "10:00 PM",
			EndTime:    "23:00",
		}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
