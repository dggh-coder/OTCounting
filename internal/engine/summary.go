package engine

import (
	"fmt"
	"sort"
	"time"
)

type dayAccumulator struct {
	rate15Minutes int
	rate20Minutes int
	rate15Segs    []string
	rate20Segs    []string
}

func formatDateLabel(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("02-Jan-2006")
}

func formatSegment(tr timeRange) string {
	return fmt.Sprintf("%s-%s", tr.Start.Format("15:04"), tr.End.Format("15:04"))
}

func sortedKeys[K ~string, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
