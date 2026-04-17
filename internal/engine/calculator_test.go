package engine

import "testing"

func TestCrossMidnightGroupedByStartDate(t *testing.T) {
	calc := NewCalculator()
	out, err := calc.Calculate(CalculateInput{
		OTEntries: []OTEntry{{ID: "1", EmployeeID: EmployeeA, Date: "2026-04-14", Period: "PM", StartTime: "23:30", EndTime: "01:00"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	day, ok := out.DailySummary[EmployeeA]["2026041402"]
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
		OTEntries:    []OTEntry{{ID: "1", EmployeeID: EmployeeA, Date: "2026-04-14", Period: "PM", StartTime: "20:00", EndTime: "21:00"}},
		BreakEntries: []BreakEntry{{ID: "b", EmployeeID: EmployeeB, Date: "2026-04-14", Period: "PM", StartTime: "20:15", EndTime: "20:45"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got := out.DailySummary[EmployeeA]["2026041402"].Rate20Minutes; got != 60 {
		t.Fatalf("expected 60 got %d", got)
	}
}

func TestSpecialShortBoundaryRuleTieGoesTo20(t *testing.T) {
	calc := NewCalculator()
	out, err := calc.Calculate(CalculateInput{
		OTEntries: []OTEntry{{ID: "1", EmployeeID: EmployeeA, Date: "2026-04-14", Period: "AM", StartTime: "06:30", EndTime: "07:30"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	day := out.DailySummary[EmployeeA]["2026041401"]
	if day.Rate20Minutes != 60 || day.Rate15Minutes != 0 {
		t.Fatalf("unexpected minutes: %+v", day)
	}
}

func TestExcludeNonOTWindow(t *testing.T) {
	calc := NewCalculator()
	out, err := calc.Calculate(CalculateInput{
		OTEntries: []OTEntry{{ID: "1", EmployeeID: EmployeeA, Date: "2026-04-14", Period: "AM", StartTime: "08:00", EndTime: "19:00"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	day := out.DailySummary[EmployeeA]["2026041401"]
	if day.Rate15Minutes != 90 { // 08:00-08:45 and 18:15-19:00
		t.Fatalf("expected 90 got %d", day.Rate15Minutes)
	}
}

func TestDuplicateOTOnlyCountedOnce(t *testing.T) {
	calc := NewCalculator()
	out, err := calc.Calculate(CalculateInput{
		OTEntries: []OTEntry{
			{ID: "1", EmployeeID: EmployeeA, Date: "2026-04-14", Period: "PM", StartTime: "20:00", EndTime: "22:00"},
			{ID: "2", EmployeeID: EmployeeA, Date: "2026-04-14", Period: "PM", StartTime: "20:00", EndTime: "22:00"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	day := out.DailySummary[EmployeeA]["2026041402"]
	if day.Rate20Minutes != 120 {
		t.Fatalf("expected 120 minutes once, got %d", day.Rate20Minutes)
	}
}

func TestMixedRoundHoursRules(t *testing.T) {
	cases := []struct {
		name                 string
		rate15Min, rate20Min int
		want15, want20       int
	}{
		{name: "sum_lt_30", rate15Min: 70, rate20Min: 79, want15: 1, want20: 1},
		{name: "sum_30_to_59_award_15", rate15Min: 85, rate20Min: 67, want15: 2, want20: 1},
		{name: "sum_30_to_59_award_20_tie", rate15Min: 75, rate20Min: 75, want15: 1, want20: 2},
		{name: "sum_ge_60_15_wins_and_20_extra", rate15Min: 110, rate20Min: 100, want15: 2, want20: 2},
		{name: "sum_ge_60_20_wins_and_15_no_extra", rate15Min: 80, rate20Min: 105, want15: 1, want20: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got15, got20 := mixedRoundHours(tc.rate15Min, tc.rate20Min)
			if got15 != tc.want15 || got20 != tc.want20 {
				t.Fatalf("got (%d,%d), want (%d,%d)", got15, got20, tc.want15, tc.want20)
			}
		})
	}
}
