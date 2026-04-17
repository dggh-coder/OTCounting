package engine

type EmployeeID string

const (
	EmployeeA EmployeeID = "A"
	EmployeeB EmployeeID = "B"
)

type OTEntry struct {
	ID         string     `json:"id"`
	EmployeeID EmployeeID `json:"employeeId"`
	Date       string     `json:"date"`
	Period     string     `json:"period"`
	StartTime  string     `json:"startTime"`
	EndTime    string     `json:"endTime"`
}

type BreakEntry struct {
	ID         string     `json:"id"`
	EmployeeID EmployeeID `json:"employeeId"`
	Date       string     `json:"date"`
	Period     string     `json:"period"`
	StartTime  string     `json:"startTime"`
	EndTime    string     `json:"endTime"`
}

type CalculateInput struct {
	OTEntries    []OTEntry    `json:"otEntries"`
	BreakEntries []BreakEntry `json:"breakEntries"`
}

type DailySummary struct {
	DateLabel          string   `json:"dateLabel"`
	Rate20Segments     []string `json:"rate20Segments"`
	Rate20Minutes      int      `json:"rate20Minutes"`
	Rate20RoundedHours int      `json:"rate20RoundedHours"`
	Rate15Segments     []string `json:"rate15Segments"`
	Rate15Minutes      int      `json:"rate15Minutes"`
	Rate15RoundedHours int      `json:"rate15RoundedHours"`
	TotalWeighted      float64  `json:"totalWeighted"`
}

type MonthlySummary struct {
	Rate15RoundedHours int     `json:"rate15RoundedHours"`
	Rate20RoundedHours int     `json:"rate20RoundedHours"`
	TotalWeighted      float64 `json:"totalWeighted"`
}

type CalculateOutput struct {
	DailySummary   map[EmployeeID]map[string]DailySummary   `json:"dailySummary"`
	MonthlySummary map[EmployeeID]map[string]MonthlySummary `json:"monthlySummary"`
}
