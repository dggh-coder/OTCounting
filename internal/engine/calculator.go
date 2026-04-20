package engine

import (
	"fmt"
	"sort"
	"time"
)

const (
	dateLayout     = "2006-01-02"
	dateTimeLayout = "2006-01-02 15:04"
)

type Calculator struct{}

func NewCalculator() *Calculator { return &Calculator{} }

func (c *Calculator) Calculate(input CalculateInput) (CalculateOutput, error) {
	breaksByEmployee := map[EmployeeID][]timeRange{}
	for _, br := range input.BreakEntries {
		r, err := parseRange(br.Date, br.StartTime, br.EndTime)
		if err != nil {
			return CalculateOutput{}, fmt.Errorf("invalid break entry %s: %w", br.ID, err)
		}
		breaksByEmployee[br.EmployeeID] = append(breaksByEmployee[br.EmployeeID], r)
	}
	for emp := range breaksByEmployee {
		sort.Slice(breaksByEmployee[emp], func(i, j int) bool {
			return breaksByEmployee[emp][i].Start.Before(breaksByEmployee[emp][j].Start)
		})
	}

	dailyAcc := map[EmployeeID]map[string]*dayAccumulator{EmployeeA: {}, EmployeeB: {}}
	seenOT := map[string]struct{}{}

	for _, ot := range input.OTEntries {
		otRange, err := parseRange(ot.Date, ot.StartTime, ot.EndTime)
		if err != nil {
			return CalculateOutput{}, fmt.Errorf("invalid ot entry %s: %w", ot.ID, err)
		}

		otKey := fmt.Sprintf("%s|%s|%s", ot.EmployeeID, otRange.Start.Format(time.RFC3339), otRange.End.Format(time.RFC3339))
		if _, exists := seenOT[otKey]; exists {
			continue
		}
		seenOT[otKey] = struct{}{}

		segments := []timeRange{otRange}
		segments = subtractNonOT(segments)
		for _, br := range breaksByEmployee[ot.EmployeeID] {
			segments = subtractRange(segments, br)
		}

		period := normalizePeriod(ot.Period, otRange.Start)
		dateKey := sessionKey(otRange.Start.Format(dateLayout), period)
		if _, ok := dailyAcc[ot.EmployeeID][dateKey]; !ok {
			dailyAcc[ot.EmployeeID][dateKey] = &dayAccumulator{
				dateLabel: formatDateLabel(otRange.Start.Format(dateLayout)) + " " + period,
			}
		}
		acc := dailyAcc[ot.EmployeeID][dateKey]

		for _, seg := range segments {
			if !seg.End.After(seg.Start) {
				continue
			}
			rate15Mins, rate20Mins := splitSegmentMinutes(seg)
			totalMinutes := int(seg.End.Sub(seg.Start).Minutes())
			if totalMinutes <= 89 && rate15Mins > 0 && rate20Mins > 0 {
				if rate20Mins >= rate15Mins {
					rate20Mins = totalMinutes
					rate15Mins = 0
				} else {
					rate15Mins = totalMinutes
					rate20Mins = 0
				}
			}

			if rate15Mins > 0 {
				acc.rate15Minutes += rate15Mins
				acc.rate15Segs = append(acc.rate15Segs, formatSegment(seg))
			}
			if rate20Mins > 0 {
				acc.rate20Minutes += rate20Mins
				acc.rate20Segs = append(acc.rate20Segs, formatSegment(seg))
			}
		}
	}

	output := CalculateOutput{
		DailySummary:   map[EmployeeID]map[string]DailySummary{EmployeeA: {}, EmployeeB: {}},
		MonthlySummary: map[EmployeeID]map[string]MonthlySummary{EmployeeA: {}, EmployeeB: {}},
	}

	for _, emp := range []EmployeeID{EmployeeA, EmployeeB} {
		monthAcc := map[string]MonthlySummary{}
		for _, dateKey := range sortedKeys(dailyAcc[emp]) {
			acc := dailyAcc[emp][dateKey]
			rate15Hours, rate20Hours := mixedRoundHours(acc.rate15Minutes, acc.rate20Minutes)
			totalWeighted := float64(rate15Hours)*1.5 + float64(rate20Hours)*2.0

			ds := DailySummary{
				DateLabel:          acc.dateLabel,
				Rate20Segments:     acc.rate20Segs,
				Rate20Minutes:      acc.rate20Minutes,
				Rate20RoundedHours: rate20Hours,
				Rate15Segments:     acc.rate15Segs,
				Rate15Minutes:      acc.rate15Minutes,
				Rate15RoundedHours: rate15Hours,
				TotalWeighted:      totalWeighted,
			}
			output.DailySummary[emp][dateKey] = ds

			monthKey := dateKey[:6]
			monthKey = monthKey[:4] + "-" + monthKey[4:]
			ms := monthAcc[monthKey]
			ms.Rate15RoundedHours += rate15Hours
			ms.Rate20RoundedHours += rate20Hours
			ms.TotalWeighted += totalWeighted
			monthAcc[monthKey] = ms
		}
		for m, ms := range monthAcc {
			output.MonthlySummary[emp][m] = ms
		}
	}

	return output, nil
}

func parseRange(dateStr, start, end string) (timeRange, error) {
	startDT, err := time.Parse(dateTimeLayout, dateStr+" "+start)
	if err != nil {
		return timeRange{}, err
	}
	endDT, err := time.Parse(dateTimeLayout, dateStr+" "+end)
	if err != nil {
		return timeRange{}, err
	}
	if !endDT.After(startDT) {
		endDT = endDT.Add(24 * time.Hour)
	}
	return timeRange{Start: startDT, End: endDT}, nil
}

func subtractNonOT(segments []timeRange) []timeRange {
	out := segments
	for _, seg := range segments {
		startDay := time.Date(seg.Start.Year(), seg.Start.Month(), seg.Start.Day(), 0, 0, 0, 0, seg.Start.Location())
		for i := -1; i <= 2; i++ {
			day := startDay.AddDate(0, 0, i)
			nonOTStart := time.Date(day.Year(), day.Month(), day.Day(), 8, 45, 0, 0, day.Location())
			nonOTEnd := time.Date(day.Year(), day.Month(), day.Day(), 18, 15, 0, 0, day.Location())
			out = subtractRange(out, timeRange{Start: nonOTStart, End: nonOTEnd})
		}
	}
	return out
}

func splitSegmentMinutes(seg timeRange) (int, int) {
	rate15Mins := 0
	rate20Mins := 0
	for cur := seg.Start; cur.Before(seg.End); cur = cur.Add(time.Minute) {
		switch classifyMinute(cur) {
		case rate15:
			rate15Mins++
		case rate20:
			rate20Mins++
		}
	}
	return rate15Mins, rate20Mins
}

func mixedRoundHours(rate15Minutes, rate20Minutes int) (int, int) {
	rate15Hours := rate15Minutes / 60
	rate20Hours := rate20Minutes / 60
	rate15MM := rate15Minutes % 60
	rate20MM := rate20Minutes % 60
	totalMM := rate15MM + rate20MM

	if totalMM < 30 {
		return rate15Hours, rate20Hours
	}

	if totalMM < 60 {
		if rate15MM > rate20MM {
			rate15Hours++
		} else {
			rate20Hours++
		}
		return rate15Hours, rate20Hours
	}

	if rate15MM > rate20MM {
		rate15Hours++
		leftover20 := rate20MM - (60 - rate15MM)
		if leftover20 >= 30 {
			rate20Hours++
		}
		return rate15Hours, rate20Hours
	}

	rate20Hours++
	leftover15 := rate15MM - (60 - rate20MM)
	if leftover15 >= 30 {
		rate15Hours++
	}
	return rate15Hours, rate20Hours
}

func normalizePeriod(raw string, start time.Time) string {
	switch raw {
	case "AM", "am", "Am", "aM":
		return "AM"
	case "PM", "pm", "Pm", "pM":
		return "PM"
	default:
		if start.Hour() < 12 {
			return "AM"
		}
		return "PM"
	}
}

func sessionKey(dateStr, period string) string {
	t, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return dateStr + "_" + period
	}
	suffix := "01"
	if period == "PM" {
		suffix = "02"
	}
	return t.Format("20060102") + suffix
}
