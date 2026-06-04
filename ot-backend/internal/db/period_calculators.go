package db

import (
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
	return calculatePeriodResult(otRanges, breakRanges)
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
