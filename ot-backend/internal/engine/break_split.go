package engine

import "time"

type timeRange struct {
	Start time.Time
	End   time.Time
}

func subtractRange(base []timeRange, sub timeRange) []timeRange {
	result := make([]timeRange, 0, len(base)+1)
	for _, r := range base {
		if !overlaps(r, sub) {
			result = append(result, r)
			continue
		}
		if sub.Start.After(r.Start) {
			result = append(result, timeRange{Start: r.Start, End: minTime(sub.Start, r.End)})
		}
		if sub.End.Before(r.End) {
			result = append(result, timeRange{Start: maxTime(sub.End, r.Start), End: r.End})
		}
	}
	return compactValid(result)
}

func overlaps(a, b timeRange) bool {
	return a.Start.Before(b.End) && b.Start.Before(a.End)
}

func compactValid(ranges []timeRange) []timeRange {
	out := ranges[:0]
	for _, r := range ranges {
		if r.End.After(r.Start) {
			out = append(out, r)
		}
	}
	return out
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
