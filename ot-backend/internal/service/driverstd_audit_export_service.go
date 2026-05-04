package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"

	"ot-uat/internal/db"
)

func BuildDriverAuditReportCSV(staffName, startDate, endDate string, rows []db.DriverMonthlyReportRow, summaryRows []db.DriverAuditSummaryRow) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.Write([]string{fmt.Sprintf("%s %s ~ %s :", staffName, startDate, endDate)}); err != nil {
		return nil, err
	}
	if err := w.Write([]string{"Date", "Start Time", "End Time"}); err != nil {
		return nil, err
	}
	detailByDate := map[string][]db.DriverMonthlyReportRow{}
	for _, r := range rows {
		detailByDate[r.Date] = append(detailByDate[r.Date], r)
	}
	byDate := map[string]map[string]db.DriverAuditSummaryRow{}
	for _, s := range summaryRows {
		if _, ok := byDate[s.Date]; !ok {
			byDate[s.Date] = map[string]db.DriverAuditSummaryRow{}
		}
		byDate[s.Date][s.Period] = s
	}

	dates := make([]string, 0, len(detailByDate))
	for d := range detailByDate {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	for _, date := range dates {
		for _, r := range detailByDate[date] {
			if err := w.Write([]string{r.Date, r.StartTime, r.EndTime}); err != nil {
				return nil, err
			}
		}
		m := byDate[date]
		p20 := fmt.Sprintf("[%s] + [%s] + [%s]", m["00"].Process20Txt, m["01"].Process20Txt, m["02"].Process20Txt)
		p15 := fmt.Sprintf("[%s] + [%s] + [%s]", m["00"].Process15Txt, m["01"].Process15Txt, m["02"].Process15Txt)
		t20 := m["00"].TotalHrs20 + m["01"].TotalHrs20 + m["02"].TotalHrs20
		t15 := m["00"].TotalHrs15 + m["01"].TotalHrs15 + m["02"].TotalHrs15
		if err := w.Write([]string{fmt.Sprintf("2.0 OT: %s = %d hrs", p20, t20)}); err != nil {
			return nil, err
		}
		if err := w.Write([]string{fmt.Sprintf("1.5 OT: %s = %d hrs", p15, t15)}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
