package engine

import "time"

type rateGroup string

const (
	rateNone rateGroup = "none"
	rate15   rateGroup = "1.5"
	rate20   rateGroup = "2.0"
)

func classifyMinute(t time.Time) rateGroup {
	h, m, _ := t.Clock()
	minuteOfDay := h*60 + m

	switch {
	case minuteOfDay >= 7*60 && minuteOfDay < 8*60+45:
		return rate15
	case minuteOfDay >= 8*60+45 && minuteOfDay < 18*60+15:
		return rateNone
	case minuteOfDay >= 18*60+15 && minuteOfDay < 20*60:
		return rate15
	default:
		return rate20
	}
}
