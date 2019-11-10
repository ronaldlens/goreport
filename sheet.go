package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"log"
	"strconv"
)

type Sheet struct {
	filename string
	file     *excelize.File
}

func (sheet *Sheet) init() {
	sheet.file = excelize.NewFile()
}

func (sheet *Sheet) wop() {
	fmt.Println("wop")
}

func (sheet *Sheet) setupExcelFile() {
	xls := sheet.file
	xls.SetActiveSheet(xls.NewSheet("Overview"))

	percentStyle, _ := xls.NewStyle(`{"number_format": 9}`)

	_ = xls.SetCellStr("Overview", "A2", "Total Incidents")
	_ = xls.SetCellStr("Overview", "A3", "Priority")
	_ = xls.SetCellStr("Overview", "A9", "SLA Met Incidents")
	_ = xls.SetCellStr("Overview", "A10", "Priority")
	_ = xls.SetCellStr("Overview", "A16", "SLA Performance")
	_ = xls.SetCellStr("Overview", "A17", "Priority")
	_ = xls.SetCellStr("Overview", "B17", "Target")

	for idx, priorityName := range priorityNames {
		axis, _ := excelize.CoordinatesToCellName(1, idx+4)
		_ = xls.SetCellStr("Overview", axis, priorityName)
		axis, _ = excelize.CoordinatesToCellName(1, idx+11)
		_ = xls.SetCellStr("Overview", axis, priorityName)
		axis, _ = excelize.CoordinatesToCellName(1, idx+18)
		_ = xls.SetCellStr("Overview", axis, priorityName)

		axis, _ = excelize.CoordinatesToCellName(2, idx+18)
		_ = xls.SetCellFloat("Overview", axis, 0.8, 2, 32)
		_ = xls.SetCellStyle("Overview", axis, axis, percentStyle)
	}
}

func (sheet *Sheet) addProdCategoriesToSheet(incidents []Incident) {
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
	categories := collectProdCategories(incidents)
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

	// setup the header row
	_ = xls.SetCellStr("Incidents", "A1", "ID")
	_ = xls.SetCellStr("Incidents", "B1", "Created")
	_ = xls.SetCellStr("Incidents", "C1", "Solved")
	_ = xls.SetCellStr("Incidents", "D1", "Time Open")
	_ = xls.SetCellStr("Incidents", "E1", "Corrected Open")
	_ = xls.SetCellStr("Incidents", "F1", "Exclude")
	_ = xls.SetCellStr("Incidents", "G1", "Priority")
	_ = xls.SetCellStr("Incidents", "H1", "Product Category")
	_ = xls.SetCellStr("Incidents", "I1", "Service CI")
	_ = xls.SetCellStr("Incidents", "J1", "SLA Met")
	_ = xls.SetCellStr("Incidents", "K1", "Description")
	_ = xls.SetCellStr("Incidents", "L1", "Resolution")

	maxProdLen := 1
	maxCILen := 1
	maxDescLen := 1
	maxResLen := 1

	for row, incident := range incidents {
		rowStr := strconv.Itoa(row + 2)

		_ = xls.SetCellValue("Incidents", "A"+rowStr, incident.ID)
		_ = xls.SetCellValue("Incidents", "B"+rowStr, incident.CreatedAt)
		if incident.SLAReady {
			_ = xls.SetCellValue("Incidents", "C"+rowStr, incident.SolvedAt)
		}
		_ = xls.SetCellValue("Incidents", "D"+rowStr, incident.OpenTime)
		_ = xls.SetCellValue("Incidents", "E"+rowStr, incident.CorrectedTime)
		_ = xls.SetCellValue("Incidents", "F"+rowStr, incident.Exclude)
		_ = xls.SetCellValue("Incidents", "G"+rowStr, priorityNames[incident.Priority])
		_ = xls.SetCellValue("Incidents", "H"+rowStr, incident.ProdCategory)
		_ = xls.SetCellValue("Incidents", "I"+rowStr, incident.ServiceCI)
		_ = xls.SetCellValue("Incidents", "J"+rowStr, incident.SLAMet)
		_ = xls.SetCellValue("Incidents", "K"+rowStr, incident.Description)
		_ = xls.SetCellValue("Incidents", "L"+rowStr, incident.Resolution)

		if len(incident.ProdCategory) > maxProdLen {
			maxProdLen = len(incident.ProdCategory)
		}
		if len(incident.ServiceCI) > maxCILen {
			maxCILen = len(incident.ServiceCI)
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
	_ = xls.SetColWidth("Incidents", "H", "H", 0.9*float64(maxProdLen))
	_ = xls.SetColWidth("Incidents", "I", "I", 0.9*float64(maxCILen))
	_ = xls.SetColWidth("Incidents", "K", "K", 0.9*float64(maxDescLen))
	_ = xls.SetColWidth("Incidents", "L", "L", 0.9*float64(maxResLen))

	rowStr := strconv.Itoa(len(incidents) + 1)
	_ = xls.AutoFilter("Incidents", "A1", "L"+rowStr, "")
}

func (sheet *Sheet) createCharts() {
	xls := sheet.file
	series := ""
	for _, i := range []int{4, 5, 6, 7} {
		series += fmt.Sprintf("{\"name\":\"Overview!$A$%d\",\"categories\":\"Overview!$C$3:$H$3\",\"values\":\"Overview!$C%d:$H%d\"}", i, i, i)
		if i != 7 {
			series += ","
		}
	}
	cs := fmt.Sprintf("{\"type\":\"line\",\"series\":[%s],", series)
	cs += "\"format\":{\"x_scale\":1.0,\"y_scale\":1.0,\"x_offset\":15,\"y_offset\":10,\"print_obj\":true,\"lock_aspect_ratio\":false,\"locked\":false},"
	cs += "\"legend\":{\"position\":\"bottom\",\"show_legend_key\":false},"
	cs += "\"title\":{\"name\":\"Total Incidents\"},"
	cs += "\"plotarea\":{\"show_bubble_size\":false,\"show_cat_name\":false,\"show_leader_lines\":true,\"show_percent\":false,\"show_series_name\":false,\"show_val\":false},\"show_blanks_as\":\"gap\"}"

	err := xls.AddChart("Overview", "J2", cs)
	if err != nil {
		log.Fatalf("Error adding chart: %v", err)
	}

	series = ""
	for _, i := range []int{18, 19, 20, 21} {
		series += fmt.Sprintf("{\"name\":\"Overview!$A$%d\",\"categories\":\"Overview!$C$3:$H$3\",\"values\":\"Overview!$C%d:$H%d\"}", i, i, i)
		if i != 21 {
			series += ","
		}
	}
	cs = fmt.Sprintf("{\"type\":\"line\",\"series\":[%s],", series)
	cs += "\"format\":{\"x_scale\":1.0,\"y_scale\":1.0,\"x_offset\":15,\"y_offset\":10,\"print_obj\":true,\"lock_aspect_ratio\":false,\"locked\":false},"
	cs += "\"legend\":{\"position\":\"bottom\",\"show_legend_key\":false},"
	cs += "\"title\":{\"name\":\"SLA Performance\"},"
	cs += "\"plotarea\":{\"show_bubble_size\":false,\"show_cat_name\":false,\"show_leader_lines\":true,\"show_percent\":false,\"show_series_name\":false,\"show_val\":false},\"show_blanks_as\":\"gap\"}"

	err = xls.AddChart("Overview", "J18", cs)
	if err != nil {
		log.Fatalf("Error adding chart: %v", err)
	}
}

func (sheet *Sheet) SaveAs(filename string) error {
	return sheet.file.SaveAs(filename)
}
