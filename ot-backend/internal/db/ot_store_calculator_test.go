package db

import (
	"reflect"
	"testing"
)

func TestSelectPeriodCalculatorByInputDate(t *testing.T) {
	cases := []struct {
		name string
		date string
		want any
	}{
		{name: "friday", date: "2026-06-05", want: FridayCal{}},
		{name: "other_day", date: "2026-06-04", want: OtherDayCal{}},
		{name: "invalid_date_falls_back_to_other_day", date: "not-a-date", want: OtherDayCal{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := selectPeriodCalculator(tc.date)
			if reflect.TypeOf(got) != reflect.TypeOf(tc.want) {
				t.Fatalf("got %T, want %T", got, tc.want)
			}
		})
	}
}

func TestFridayCalMatchesOtherDayCalCurrentRules(t *testing.T) {
	otOne, err := parseDateRange("2026-06-05", "17:45", "19:00")
	if err != nil {
		t.Fatalf("parse first ot range: %v", err)
	}
	otTwo, err := parseDateRange("2026-06-05", "20:00", "21:30")
	if err != nil {
		t.Fatalf("parse second ot range: %v", err)
	}

	friday := FridayCal{}.Calculate([]timeSpan{otOne, otTwo}, nil)
	fridayDirect := FridayCalculatePeriodResult([]timeSpan{otOne, otTwo}, nil)
	otherDay := OtherDayCal{}.Calculate([]timeSpan{otOne, otTwo}, nil)

	if !reflect.DeepEqual(friday, fridayDirect) {
		t.Fatalf("FridayCal differs from FridayCalculatePeriodResult: friday=%+v direct=%+v", friday, fridayDirect)
	}
	if !reflect.DeepEqual(friday, otherDay) {
		t.Fatalf("FridayCal differs from OtherDayCal: friday=%+v otherDay=%+v", friday, otherDay)
	}
	if friday.rate15Mins != 45 || friday.rate20Mins != 90 {
		t.Fatalf("unexpected minutes: got 1.5=%d 2.0=%d, want 1.5=45 2.0=90", friday.rate15Mins, friday.rate20Mins)
	}
}
