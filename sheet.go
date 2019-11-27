package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"log"
	"strconv"
	"strings"
	"time"
)

// Sheet is a struct containing the filename and the excelize.File
type Sheet struct {
	filename string
	file     *excelize.File
}

func (sheet *Sheet) init() {
	sheet.file = excelize.NewFile()
}

func (sheet *Sheet) setupOverviewSheet(area string) {
	xls := sheet.file
	if area == "" {
		xls.SetActiveSheet(xls.NewSheet("Overview"))
	} else {
		area = " " + area
		xls.SetActiveSheet(xls.NewSheet("Overview" + area))
	}

	percentStyle, _ := xls.NewStyle(`{"number_format": 9}`)

	_ = xls.SetCellStr("Overview"+area, "A2", "Total Incidents"+area)
	_ = xls.SetCellStr("Overview"+area, "A3", "Priority")
	_ = xls.SetCellStr("Overview"+area, "A9", "SLA Met Incidents"+area)
	_ = xls.SetCellStr("Overview"+area, "A10", "Priority")
	_ = xls.SetCellStr("Overview"+area, "A16", "SLA Performance"+area)
	_ = xls.SetCellStr("Overview"+area, "A17", "Priority")
	_ = xls.SetCellStr("Overview"+area, "B17", "Target")

	for idx, priorityName := range PriorityNames {
		axis, _ := excelize.CoordinatesToCellName(1, idx+4)
		_ = xls.SetCellStr("Overview"+area, axis, priorityName)
		axis, _ = excelize.CoordinatesToCellName(1, idx+11)
		_ = xls.SetCellStr("Overview"+area, axis, priorityName)
		axis, _ = excelize.CoordinatesToCellName(1, idx+18)
		_ = xls.SetCellStr("Overview"+area, axis, priorityName)

		axis, _ = excelize.CoordinatesToCellName(2, idx+18)
		_ = xls.SetCellFloat("Overview"+area, axis, 0.8, 2, 32)
		_ = xls.SetCellStyle("Overview"+area, axis, axis, percentStyle)
	}
}

func (sheet *Sheet) addProdCategoriesToSheet(incidents Incidents) {
	xls := sheet.file
	xls.SetActiveSheet(xls.NewSheet("ProdCat"))

	// set the headers
	_ = xls.SetCellStr("ProdCat", "A1", "Product Category")
	_ = xls.SetCellStr("ProdCat", "B1", "Total")
	_ = xls.SetCellStr("ProdCat", "C1", "Met SLA")
	_ = xls.SetCellStr("ProdCat", "D1", "Critical")
	_ = xls.SetCellStr("ProdCat", "E1", "High")
	_ = xls.SetCellStr("ProdCat", "F1", "Medium")
	_ = xls.SetCellStr("ProdCat", "G1", "Low")

	// get categories, initialize the row
	// initialize max length to later set the width of the name column
	categories := incidents.collectProdCategories()
	row := 1
	maxLen := 1

	for categoryName, category := range categories {
		// find the maximum width
		if len(categoryName) > maxLen {
			maxLen = len(categoryName)
		}

		// increment the row and turn into string
		row++
		rowStr := strconv.Itoa(row)

		_ = xls.SetCellStr("ProdCat", "A"+rowStr, categoryName)
		_ = xls.SetCellInt("ProdCat", "B"+rowStr, category.Total)
		_ = xls.SetCellInt("ProdCat", "C"+rowStr, category.SLAMet)
		_ = xls.SetCellInt("ProdCat", "D"+rowStr, category.Critical)
		_ = xls.SetCellInt("ProdCat", "E"+rowStr, category.High)
		_ = xls.SetCellInt("ProdCat", "F"+rowStr, category.Medium)
		_ = xls.SetCellInt("ProdCat", "G"+rowStr, category.Low)
	}

	// turn on autofilter and set the width to fit the name in column A
	_ = xls.AutoFilter("ProdCat", "A1", "G"+strconv.Itoa(row), "")
	_ = xls.SetColWidth("ProdCat", "A", "A", 0.9*float64(maxLen))
}

func (sheet *Sheet) addIncidentsToSheet(incidents []Incident) {
	xls := sheet.file
	xls.SetActiveSheet(xls.NewSheet("Incidents"))
	urlPrefix := "http://usms.upc.biz/arsys/forms/appusms/SHR%%3ALandingConsole/Default+Administrator+View/" +
		"?mode=search&F304255500=HPD%%3AHelp+Desk&F1000000076=FormOpenNoAppList&F303647600=" +
		"SearchTicketWithQual&F304255610='1000000161'%%3D%%22%s%%22"
	urlStyle, _ := xls.NewStyle(`{"font":{"color":"#1265BE","underline":"single"}}`)

	// setup the header row
	_ = xls.SetCellStr("Incidents", "A1", "ID")
	_ = xls.SetCellStr("Incidents", "B1", "Created")
	_ = xls.SetCellStr("Incidents", "C1", "Solved")
	_ = xls.SetCellStr("Incidents", "D1", "Time Open")
	_ = xls.SetCellStr("Incidents", "E1", "Corrected Open")
	_ = xls.SetCellStr("Incidents", "F1", "Exclude")
	_ = xls.SetCellStr("Incidents", "G1", "Priority")
	_ = xls.SetCellStr("Incidents", "H1", "Product Category Tier 1")
	_ = xls.SetCellStr("Incidents", "I1", "Product Category Tier 2")
	_ = xls.SetCellStr("Incidents", "J1", "Service")
	_ = xls.SetCellStr("Incidents", "K1", "Service CI")
	_ = xls.SetCellStr("Incidents", "L1", "Business Area")
	_ = xls.SetCellStr("Incidents", "M1", "SLA Met")
	_ = xls.SetCellStr("Incidents", "N1", "Description")
	_ = xls.SetCellStr("Incidents", "O1", "Resolution")

	maxProdCat1Len := 1
	maxProdCat2Len := 1
	maxSvcLen := 1
	maxCILen := 1
	maxDescLen := 1
	maxResLen := 1

	for row, incident := range incidents {
		rowStr := strconv.Itoa(row + 2)

		_ = xls.SetCellValue("Incidents", "A"+rowStr, incident.ID)
		url := fmt.Sprintf(urlPrefix, incident.ID)
		_ = xls.SetCellHyperLink("Incidents", "A"+rowStr, url, "External")
		_ = xls.SetCellStyle("Incidents", "A"+rowStr, "A"+rowStr, urlStyle)

		_ = xls.SetCellValue("Incidents", "B"+rowStr, incident.CreatedAt)
		if incident.SLAReady {
			_ = xls.SetCellValue("Incidents", "C"+rowStr, incident.SolvedAt)
		}
		_ = xls.SetCellValue("Incidents", "D"+rowStr, incident.OpenTime)
		_ = xls.SetCellValue("Incidents", "E"+rowStr, incident.CorrectedTime)
		_ = xls.SetCellValue("Incidents", "F"+rowStr, incident.Exclude)
		_ = xls.SetCellValue("Incidents", "G"+rowStr, PriorityNames[incident.Priority])
		_ = xls.SetCellValue("Incidents", "H"+rowStr, incident.ProdCategory1)
		_ = xls.SetCellValue("Incidents", "I"+rowStr, incident.ProdCategory2)
		_ = xls.SetCellValue("Incidents", "J"+rowStr, incident.Service)
		_ = xls.SetCellValue("Incidents", "K"+rowStr, incident.ServiceCI)
		_ = xls.SetCellValue("Incidents", "L"+rowStr, incident.BusinessArea)
		_ = xls.SetCellValue("Incidents", "M"+rowStr, incident.SLAMet)
		_ = xls.SetCellValue("Incidents", "N"+rowStr, incident.Description)
		_ = xls.SetCellValue("Incidents", "O"+rowStr, incident.Resolution)

		if len(incident.ProdCategory1) > maxProdCat1Len {
			maxProdCat1Len = len(incident.ProdCategory1)
		}
		if len(incident.ProdCategory2) > maxProdCat2Len {
			maxProdCat2Len = len(incident.ProdCategory2)
		}
		if len(incident.ServiceCI) > maxCILen {
			maxCILen = len(incident.ServiceCI)
		}
		if len(incident.Service) > maxSvcLen {
			maxSvcLen = len(incident.Service)
		}
		if len(incident.Description) > maxDescLen {
			maxDescLen = len(incident.Description)
		}
		if len(incident.Resolution) > maxResLen {
			maxResLen = len(incident.Resolution)
		}
	}

	_ = xls.SetColWidth("Incidents", "A", "A", 16.0)
	_ = xls.SetColWidth("Incidents", "B", "B", 16.0)
	_ = xls.SetColWidth("Incidents", "C", "C", 16.0)
	_ = xls.SetColWidth("Incidents", "H", "H", 0.9*float64(maxProdCat1Len))
	_ = xls.SetColWidth("Incidents", "I", "I", 0.9*float64(maxProdCat2Len))
	_ = xls.SetColWidth("Incidents", "J", "J", 0.9*float64(maxSvcLen))
	_ = xls.SetColWidth("Incidents", "K", "K", 0.9*float64(maxCILen))
	_ = xls.SetColWidth("Incidents", "N", "M", 0.9*float64(maxDescLen))
	_ = xls.SetColWidth("Incidents", "O", "N", 0.9*float64(maxResLen))

	rowStr := strconv.Itoa(len(incidents) + 1)
	_ = xls.AutoFilter("Incidents", "A1", "M"+rowStr, "")
}

func (sheet *Sheet) createCharts(area string) {
	xls := sheet.file
	if area != "" {
		area = " " + area
	}

	series := ""
	for _, i := range []int{4, 5, 6, 7} {
		series += fmt.Sprintf("{\"name\":\"'Overview"+area+"'!$A$%d\",\"categories\":\"'Overview"+area+"'!$C$3:$H$3\",\"values\":\"'Overview"+area+"'!$C%d:$H%d\"}", i, i, i)
		if i != 7 {
			series += ","
		}
	}
	cs := fmt.Sprintf("{\"type\":\"line\",\"series\":[%s],", series)
	cs += "\"format\":{\"x_scale\":1.0,\"y_scale\":1.0,\"x_offset\":15,\"y_offset\":10,\"print_obj\":true,\"lock_aspect_ratio\":false,\"locked\":false},"
	cs += "\"legend\":{\"position\":\"bottom\",\"show_legend_key\":false},"
	cs += "\"title\":{\"name\":\"Total Incidents" + area + "\"},"
	cs += "\"plotarea\":{\"show_bubble_size\":false,\"show_cat_name\":false,\"show_leader_lines\":true,\"show_percent\":false,\"show_series_name\":false,\"show_val\":false},\"show_blanks_as\":\"gap\"}"

	err := xls.AddChart("Overview"+area, "J2", cs)
	if err != nil {
		log.Fatalf("Error adding chart: %v", err)
	}

	series = ""
	for _, i := range []int{18, 19, 20, 21} {
		series += fmt.Sprintf("{\"name\":\"'Overview"+area+"'!$A$%d\",\"categories\":\"'Overview"+area+"'!$C$3:$H$3\",\"values\":\"'Overview"+area+"'!$C%d:$H%d\"}", i, i, i)
		if i != 21 {
			series += ","
		}
	}
	cs = fmt.Sprintf("{\"type\":\"line\",\"series\":[%s],", series)
	cs += "\"format\":{\"x_scale\":1.0,\"y_scale\":1.0,\"x_offset\":15,\"y_offset\":10,\"print_obj\":true,\"lock_aspect_ratio\":false,\"locked\":false},"
	cs += "\"legend\":{\"position\":\"bottom\",\"show_legend_key\":false},"
	cs += "\"title\":{\"name\":\"SLA Performance" + area + "\"},"
	cs += "\"plotarea\":{\"show_bubble_size\":false,\"show_cat_name\":false,\"show_leader_lines\":true,\"show_percent\":false,\"show_series_name\":false,\"show_val\":false},\"show_blanks_as\":\"gap\"}"

	err = xls.AddChart("Overview"+area, "J18", cs)
	if err != nil {
		log.Fatalf("Error adding chart: %v", err)
	}
}

func getFilename(country string, month int, year int) string {
	return fmt.Sprintf("report-%s-%02d-%d.xlsx", strings.ToLower(country), month, year)
}

// SaveAs fixes the sheets (removes the default 'Sheet1' and sets the active sheet to the next one)
// and saves it. It returns an error if saving fails
func (sheet *Sheet) SaveAs(filename string) error {
	sheet.file.DeleteSheet("Sheet1")
	sheet.file.SetActiveSheet(2)
	return sheet.file.SaveAs(filename)
}

// ProcessReferenceFile processes the chnages made in an excel file.
// If a reference Excel file is provided, go through the incident sheet of that workbook
// and update our list oif incidents with ones that have a corrected outage time
// or are marked to be excluded in the reference workbook
func ProcessReferenceFile(incidents []Incident, referenceFilename string) []Incident {

	// open the reference workbook
	file, err := excelize.OpenFile(referenceFilename)
	if err != nil {
		log.Fatalf("Error opening reference file %s, %v", referenceFilename, err)
	}

	// get all the rows in the sheet titled Incidents
	rows, err := file.GetRows("Incidents")
	if err != nil {
		log.Fatalf("Error reading rows: %x", err)
	}

	// skip the first row, this is the header row
	for index, row := range rows {
		if index == 0 {
			continue
		}

		// column 4 contains the potentially updated outage time in minutes
		// if it is not 0, copy the value to our incidents list
		// if the index equals -1, the incident row could not be found
		if row[4] != "" {
			idx := findIncidentByID(incidents, row[0])
			if idx != -1 {
				incidents[idx].CorrectedTime = row[4]
				incidents[idx].CorrectedOpenTime, err = time.ParseDuration(row[4])
				if err != nil {
					log.Fatalf("Error parsing corrected time '%s' for incident %s: %v", row[4], row[0], err)
				}
			}
		}

		// column 5 contains whether an incident is to be excluded in the calculations
		if row[5] != "0" {
			idx := findIncidentByID(incidents, row[0])
			if idx != -1 {
				incidents[idx].Exclude = true
				incidents[idx].SLAReady = false
			}
		}
	}
	return incidents
}

// give an ID, find the index of the row for the incident with the ID
// if the row cannot be found, return -1
func findIncidentByID(incidents []Incident, ID string) int {
	for idx, incident := range incidents {
		if incident.ID == ID {
			return idx
		}
	}
	return -1
}
