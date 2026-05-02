package service

import "ot-uat/internal/db"

// BuildDriverMonthlyReportXLSX currently reuses CSV-compatible tabular content
// so Excel can open it directly when downloaded as .xlsx from UI.
func BuildDriverMonthlyReportXLSX(staffName, yyyymm string, rows []db.DriverMonthlyReportRow, total20, total15 int64) ([]byte, error) {
	return BuildDriverMonthlyReportCSV(staffName, yyyymm, rows, total20, total15)
}
