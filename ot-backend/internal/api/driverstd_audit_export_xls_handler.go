package api

import (
	"fmt"
	"net/http"
	"strings"

	"ot-uat/internal/service"
)

func (h *OTHandler) DriverAuditReportExportXLSX(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	staffID := strings.TrimSpace(r.URL.Query().Get("otstaffid"))
	startDate := normalizeDate(strings.TrimSpace(r.URL.Query().Get("startDate")))
	endDate := normalizeDate(strings.TrimSpace(r.URL.Query().Get("endDate")))
	staffName := strings.TrimSpace(r.URL.Query().Get("staffname"))
	if staffID == "" || startDate == "" || endDate == "" {
		http.Error(w, "otstaffid, startDate and endDate are required", http.StatusBadRequest)
		return
	}
	if staffName == "" {
		staffName = staffID
	}
	detail, summary, err := h.Store.GetDriverAuditReportRows(r.Context(), staffID, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	xlsxBytes, err := service.BuildDriverAuditReportXLSX(staffName, startDate, endDate, detail, summary)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=driver_audit_report_%s_%s_%s.xlsx", staffID, startDate, endDate))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(xlsxBytes)
}
