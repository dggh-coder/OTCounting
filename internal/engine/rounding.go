package engine

func roundMergedMinutesToHours(minutes int) int {
	if minutes <= 0 {
		return 0
	}
	return (minutes + 30) / 60
}
