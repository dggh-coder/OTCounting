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

func TestFridayCalUsesFridayEveningStart(t *testing.T) {
	ot, err := parseDateRange("2026-06-05", "17:45", "18:15")
	if err != nil {
		t.Fatalf("parse ot range: %v", err)
	}

	friday := FridayCal{}.Calculate([]timeSpan{ot}, nil)
	fridayDirect := FridayCalculatePeriodResult([]timeSpan{ot}, nil)
	otherDay := OtherDayCal{}.Calculate([]timeSpan{ot}, nil)

	if !reflect.DeepEqual(friday, fridayDirect) {
		t.Fatalf("FridayCal differs from FridayCalculatePeriodResult: friday=%+v direct=%+v", friday, fridayDirect)
	}
	if friday.rate15Mins != 30 || friday.rate20Mins != 0 {
		t.Fatalf("FridayCal got 1.5=%d 2.0=%d, want 1.5=30 2.0=0", friday.rate15Mins, friday.rate20Mins)
	}
	if !reflect.DeepEqual(friday.rate15Parts, []string{"(17:45-18:15)"}) {
		t.Fatalf("FridayCal rate15Parts=%v, want [(17:45-18:15)]", friday.rate15Parts)
	}
	if otherDay.rate15Mins != 0 || otherDay.rate20Mins != 0 {
		t.Fatalf("OtherDayCal got 1.5=%d 2.0=%d, want 1.5=0 2.0=0", otherDay.rate15Mins, otherDay.rate20Mins)
	}
}

func TestFridayCalKeepsEveningRate15Through2000(t *testing.T) {
	ot, err := parseDateRange("2026-06-05", "17:45", "20:00")
	if err != nil {
		t.Fatalf("parse ot range: %v", err)
	}

	friday := FridayCal{}.Calculate([]timeSpan{ot}, nil)
	if friday.rate15Mins != 135 || friday.rate20Mins != 0 {
		t.Fatalf("FridayCal got 1.5=%d 2.0=%d, want 1.5=135 2.0=0", friday.rate15Mins, friday.rate20Mins)
	}
	if !reflect.DeepEqual(friday.rate15Parts, []string{"(17:45-20:00)"}) {
		t.Fatalf("FridayCal rate15Parts=%v, want [(17:45-20:00)]", friday.rate15Parts)
	}
}
