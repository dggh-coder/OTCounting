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
	for _, r := range rows {
		if err := w.Write([]string{r.Date, r.StartTime, r.EndTime}); err != nil {
			return nil, err
		}
	}
	if err := w.Write([]string{fmt.Sprintf("%s %s; 2.0 Total: %d, 1.5 Total: %d", staffName, yyyymm, total20, total15)}); err != nil {
		return nil, err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
