package service

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"ot-uat/internal/db"
)

func BuildDriverMonthlyReportCSV(staffName, yyyymm string, rows []db.DriverMonthlyReportRow, total20, total15 int64) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.Write([]string{fmt.Sprintf("%s %s :", staffName, yyyymm)}); err != nil {
		return nil, err
	}
	if err := w.Write([]string{"Date", "Start Time", "End Time"}); err != nil {
		return nil, err
	}
	lastDate := ""
	for _, r := range rows {
		if r.Date != lastDate {
			if err := w.Write([]string{fmt.Sprintf("%s Justification: %s", r.Date, r.Remarks)}); err != nil {
				return nil, err
			}
			lastDate = r.Date
		}
		if err := w.Write([]string{r.Date, r.StartTime, r.EndTime}); err != nil {
			return nil, err
		}
	}
	if err := w.Write([]string{fmt.Sprintf("2.0 Total: %d hrs", total20)}); err != nil {
		return nil, err
	}
	if err := w.Write([]string{fmt.Sprintf("1.5 Total: %d hrs", total15)}); err != nil {
		return nil, err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
