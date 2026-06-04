package db

import (
	"fmt"
	"strings"
	"time"
)

type periodCalculationResult struct {
	rate15Parts []string
	rate20Parts []string
	rate15Mins  int
	rate20Mins  int
}

type periodCalculator interface {
	Calculate(otRanges, breakRanges []timeSpan) periodCalculationResult
}

type OtherDayCal struct{}

type FridayCal struct{}

func (OtherDayCal) Calculate(otRanges, breakRanges []timeSpan) periodCalculationResult {
	return calculatePeriodResult(otRanges, breakRanges)
}

func (FridayCal) Calculate(otRanges, breakRanges []timeSpan) periodCalculationResult {
	return FridayCalculatePeriodResult(otRanges, breakRanges)
}

func selectPeriodCalculator(date string) periodCalculator {
	day, err := time.Parse("2006-01-02", strings.TrimSpace(date))
	if err == nil && day.Weekday() == time.Friday {
		return FridayCal{}
	}
	return OtherDayCal{}
}

func calculatePeriodResult(otRanges, breakRanges []timeSpan) periodCalculationResult {
	out := periodCalculationResult{
		rate15Parts: []string{},
		rate20Parts: []string{},
	}

	for _, ot := range otRanges {
		segments := []timeSpan{ot}
		for _, br := range breakRanges {
			segments = subtractTmRange(segments, br)
		}
		for _, seg := range segments {
			if !seg.end.After(seg.start) {
				continue
			}
			cur := seg.start
			for cur.Before(seg.end) {
				next := cur.Add(time.Minute)
				rate := classifyRate(cur)
				if rate == 15 {
					out.rate15Mins++
				} else if rate == 20 {
					out.rate20Mins++
				}
				cur = next
			}
			r15Segs, r20Segs := splitSegmentsByRate(seg)
			out.rate15Parts = append(out.rate15Parts, r15Segs...)
			out.rate20Parts = append(out.rate20Parts, r20Segs...)
		}
	}

	return out
}

func FridayCalculatePeriodResult(otRanges, breakRanges []timeSpan) periodCalculationResult {
	out := periodCalculationResult{
		rate15Parts: []string{},
		rate20Parts: []string{},
	}

	for _, ot := range otRanges {
		segments := []timeSpan{ot}
		for _, br := range breakRanges {
			segments = subtractTmRange(segments, br)
		}
		for _, seg := range segments {
			if !seg.end.After(seg.start) {
				continue
			}
			cur := seg.start
			for cur.Before(seg.end) {
				next := cur.Add(time.Minute)
				rate := classifyFridayRate(cur)
				if rate == 15 {
					out.rate15Mins++
				} else if rate == 20 {
					out.rate20Mins++
				}
				cur = next
			}
			r15Segs, r20Segs := splitFridaySegmentsByRate(seg)
			out.rate15Parts = append(out.rate15Parts, r15Segs...)
			out.rate20Parts = append(out.rate20Parts, r20Segs...)
		}
	}

	return out
}

func splitFridaySegmentsByRate(seg timeSpan) ([]string, []string) {
	r15 := []string{}
	r20 := []string{}
	if !seg.end.After(seg.start) {
		return r15, r20
	}
	curStart := seg.start
	curRate := classifyFridayRate(seg.start)
	for cur := seg.start.Add(time.Minute); !cur.After(seg.end); cur = cur.Add(time.Minute) {
		if cur.Equal(seg.end) || classifyFridayRate(cur) != curRate {
			part := fmt.Sprintf("(%s-%s)", curStart.Format("15:04"), cur.Format("15:04"))
			if curRate == 15 {
				r15 = append(r15, part)
			} else if curRate == 20 {
				r20 = append(r20, part)
			}
			curStart = cur
			if !cur.Equal(seg.end) {
				curRate = classifyFridayRate(cur)
			}
		}
	}
	return r15, r20
}

func classifyFridayRate(t time.Time) int {
	mins := t.Hour()*60 + t.Minute()
	switch {
	case mins >= 7*60 && mins < 8*60+45:
		return 15
	case mins >= 13*60 && mins < 14*60:
		return 15
	case mins >= 17*60+45 && mins < 20*60:
		return 15
	case mins >= 8*60+45 && mins < 13*60:
		return 0
	case mins >= 14*60 && mins < 17*60+45:
		return 0
	default:
		return 20
	}
}
