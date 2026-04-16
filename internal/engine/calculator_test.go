package engine

import "testing"

func TestCrossMidnightGroupedByStartDate(t *testing.T) {
	calc := NewCalculator()
	out, err := calc.Calculate(CalculateInput{
		OTEntries: []OTEntry{{ID: "1", EmployeeID: EmployeeA, Date: "2026-04-14", StartTime: "23:30", EndTime: "01:00"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	day, ok := out.DailySummary[EmployeeA]["2026-04-14"]
	if !ok {
		t.Fatalf("expected start date key")
	}
	if day.Rate20Minutes != 90 {
		t.Fatalf("got %d", day.Rate20Minutes)
	}
}

func TestBreaksOnlySameEmployee(t *testing.T) {
	calc := NewCalculator()
	out, err := calc.Calculate(CalculateInput{
		OTEntries:    []OTEntry{{ID: "1", EmployeeID: EmployeeA, Date: "2026-04-14", StartTime: "20:00", EndTime: "21:00"}},
		BreakEntries: []BreakEntry{{ID: "b", EmployeeID: EmployeeB, Date: "2026-04-14", StartTime: "20:15", EndTime: "20:45"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got := out.DailySummary[EmployeeA]["2026-04-14"].Rate20Minutes; got != 60 {
		t.Fatalf("expected 60 got %d", got)
	}
}

func TestSpecialShortBoundaryRuleTieGoesTo20(t *testing.T) {
	calc := NewCalculator()
	out, err := calc.Calculate(CalculateInput{
		OTEntries: []OTEntry{{ID: "1", EmployeeID: EmployeeA, Date: "2026-04-14", StartTime: "06:30", EndTime: "07:30"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	day := out.DailySummary[EmployeeA]["2026-04-14"]
	if day.Rate20Minutes != 60 || day.Rate15Minutes != 0 {
		t.Fatalf("unexpected minutes: %+v", day)
	}
}

func TestExcludeNonOTWindow(t *testing.T) {
	calc := NewCalculator()
	out, err := calc.Calculate(CalculateInput{
		OTEntries: []OTEntry{{ID: "1", EmployeeID: EmployeeA, Date: "2026-04-14", StartTime: "08:00", EndTime: "19:00"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	day := out.DailySummary[EmployeeA]["2026-04-14"]
	if day.Rate15Minutes != 90 { // 08:00-08:45 and 18:15-19:00
		t.Fatalf("expected 90 got %d", day.Rate15Minutes)
	}
}
