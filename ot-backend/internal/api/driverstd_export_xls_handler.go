package api

import (
	"fmt"
	"net/http"
	"strings"

	"ot-uat/internal/service"
)

func (h *OTHandler) DriverMonthlyReportExportXLSX(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	staffID := strings.TrimSpace(r.URL.Query().Get("otstaffid"))
	yyyymm := strings.TrimSpace(r.URL.Query().Get("yyyymm"))
	staffName := strings.TrimSpace(r.URL.Query().Get("staffname"))
	if staffID == "" || yyyymm == "" {
		http.Error(w, "otstaffid and yyyymm are required", http.StatusBadRequest)
		return
	}
	if staffName == "" {
		staffName = staffID
	}
	rows, err := h.Store.GetDriverMonthlyReportRows(r.Context(), staffID, yyyymm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	summaryRows, err := h.Store.GetDriverMonthlySummary(r.Context(), yyyymm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	total20, total15 := int64(0), int64(0)
	for _, s := range summaryRows {
		if strings.EqualFold(strings.TrimSpace(s.OTStaffID), strings.TrimSpace(staffID)) {
			total20 = s.TotalHrs20
			total15 = s.TotalHrs15
			break
		}
	}
	xlsxBytes, err := service.BuildDriverMonthlyReportXLSX(staffName, yyyymm, rows, total20, total15)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=driver_monthly_report_%s_%s.xlsx", staffID, yyyymm))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(xlsxBytes)
}
