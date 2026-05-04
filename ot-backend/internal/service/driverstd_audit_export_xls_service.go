package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"

	"ot-uat/internal/db"
)

func BuildDriverAuditReportXLSX(staffName, startDate, endDate string, rows []db.DriverMonthlyReportRow, summaryRows []db.DriverAuditSummaryRow) ([]byte, error) {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	write := func(name, content string) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(content))
		return err
	}
	esc := func(s string) string { var b bytes.Buffer; _ = xml.EscapeText(&b, []byte(s)); return b.String() }

	sheetRows := []string{rowXML(1, esc(fmt.Sprintf("%s %s ~ %s :", staffName, startDate, endDate))), headerRowXML(2, esc("Date"), esc("Start Time"), esc("End Time"))}
	rowNo := 3
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
	for _, d := range dates {
		for _, r := range detailByDate[d] {
			sheetRows = append(sheetRows, dataRowXML(rowNo, esc(r.Date), esc(r.StartTime), esc(r.EndTime)))
			rowNo++
		}
		m := byDate[d]
		p20 := fmt.Sprintf("[%s] + [%s] + [%s]", m["00"].Process20Txt, m["01"].Process20Txt, m["02"].Process20Txt)
		p15 := fmt.Sprintf("[%s] + [%s] + [%s]", m["00"].Process15Txt, m["01"].Process15Txt, m["02"].Process15Txt)
		t20 := m["00"].TotalHrs20 + m["01"].TotalHrs20 + m["02"].TotalHrs20
		t15 := m["00"].TotalHrs15 + m["01"].TotalHrs15 + m["02"].TotalHrs15
		sheetRows = append(sheetRows, rowXML(rowNo, esc(fmt.Sprintf("2.0 OT: %s = %d hrs", p20, t20))))
		rowNo++
		sheetRows = append(sheetRows, rowXML(rowNo, esc(fmt.Sprintf("1.5 OT: %s = %d hrs", p15, t15))))
		rowNo++
	}

	sheetXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <dimension ref="A1:C` + fmt.Sprintf("%d", rowNo) + `"/>
  <sheetViews><sheetView workbookViewId="0"/></sheetViews>
  <sheetFormatPr defaultRowHeight="15"/>
  <sheetData>
    ` + strings.Join(sheetRows, "\n    ") + `
  </sheetData>
</worksheet>`

	if err := write("[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`); err != nil {
		return nil, err
	}
	if err := write("_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`); err != nil {
		return nil, err
	}
	if err := write("xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Audit" sheetId="1" r:id="rId1"/></sheets></workbook>`); err != nil {
		return nil, err
	}
	if err := write("xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`); err != nil {
		return nil, err
	}
	if err := write("xl/worksheets/sheet1.xml", sheetXML); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
