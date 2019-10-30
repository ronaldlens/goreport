package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/spf13/pflag"
)

const (
	Critical = iota
	High
	Medium
	Low
)

var priorityNames = []string{"Critical", "High", "Medium", "Low"}

type Incident struct {
	Country      string
	ID           string
	CreatedAt    time.Time
	SolvedAt     time.Time
	Priority     int
	Description  string
	Resolution   string
	Service      string
	ProdCategory string
	ServiceCI    string
	BusinessArea string
	SLAReady     bool
	SLAMet       bool
	OpenTime     int
}

type ProdCategory struct {
	Total    int
	Critical int
	High     int
	Medium   int
	Low      int
	SLAMet   int
}

func main() {
	start := time.Now()

	flag.String("f", "", "filename to load")
	flag.Int("m", -1, "month to retrieve")
	flag.Int("y", -1, "year to retrieve")
	flag.String("c", "Austria", "Country to retrieve")
	flag.String("x", "out.xlsx", "Excel file to write")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatalf("Error parsing commandline options: %v", err)
	}

	filename := viper.GetString("f")
	country := viper.GetString("c")
	month := viper.GetInt("m")
	year := viper.GetInt("y")
	xlsFilename := viper.GetString("x")

	incidents := readFile(filename)

	log.Printf("Loaded a total of %d incidents from %s\n", len(incidents), filename)

	incidents = filterByCountry(incidents, country)
	incidents = filterOutProdCategories(incidents, getProdCategoriesToExclude())

	xls := excelize.NewFile()
	xls.SetActiveSheet(xls.NewSheet("Overview"))
	setupExcelFile(*xls)

	if month == -1 || year == -1 {
		month = int(time.Now().Month())
		year = time.Now().Year()
		month, year = getPreviousMonth(month, year)
	}

	reportOnSixMonths(incidents, month, year, *xls)
	createCharts(*xls)

	err = xls.SaveAs(xlsFilename)
	if err != nil {
		log.Fatalf("Error saving excel file: %v", err)
	}
	log.Printf("Wrote output to %s", xlsFilename)
	log.Printf("Total running time: %s\n", time.Since(start))
}

func readFile(filename string) []Incident {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Trying to open %s: %v", filename, err)
	}

	scanner := bufio.NewScanner(transform.NewReader(
		file, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))

	// skip the firsl line, header
	scanner.Scan()

	var incidents []Incident
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		var inc = Incident{
			Country:      parts[0],
			ID:           parts[4],
			Description:  parts[14],
			Resolution:   parts[15],
			ProdCategory: parts[9],
			Service:      parts[10],
			ServiceCI:    parts[11],
			BusinessArea: parts[12],
		}

		// get timestamps, createdAt and solvedAt
		// solvedAt may not be filled in (yet), if so, don't use it for SLA calculations
		t, err := parseTimeStamp(parts[3])
		if err == nil {
			inc.CreatedAt = t
		}
		t, err = parseTimeStamp(parts[4])
		if err == nil {
			inc.SolvedAt = t
			inc.SLAReady = true
		}

		switch parts[6] {
		case "Critical":
			inc.Priority = Critical
		case "High":
			inc.Priority = High
		case "Medium":
			inc.Priority = Medium
		case "Low":
			inc.Priority = Low
		}

		// calculate the open time in minutes
		// check if the SLA is met
		if inc.SLAReady {
			inc.OpenTime = int(inc.SolvedAt.Sub(inc.CreatedAt).Minutes())
		} else {
			inc.OpenTime = 0
		}
		inc.SLAMet = checkSLA(&inc)

		incidents = append(incidents, inc)
	}
	err = file.Close()
	if err != nil {
		log.Fatalf("Error closing file %s: %v", filename, err)
	}
	return incidents
}

func parseTimeStamp(ts string) (time.Time, error) {
	ts = strings.Replace(ts, " ", "T", 1) // Make it RFC3339 compliant
	ts = strings.Replace(ts, "/", "-", 2) // fix the date part
	ts += "Z"                             // add the Z for UTC
	return time.Parse(time.RFC3339, ts)
}

func checkSLA(incident *Incident) bool {
	// only process if the incident is solved
	if !incident.SLAReady {
		return false
	}
	switch incident.Priority {
	case Critical:
		return checkSLAHours(incident, 4)
	case High:
		return checkSLAHours(incident, 8)
	case Medium:
		return checkSLABusinessDays(incident, 1)
	case Low:
		return checkSLABusinessDays(incident, 3)
	}
	return false
}

func checkSLAHours(incident *Incident, hours int) bool {
	duration, _ := time.ParseDuration(fmt.Sprintf("%dh", hours))
	target := incident.CreatedAt.Add(duration)
	return target.After(incident.SolvedAt)
}

func checkSLABusinessDays(incident *Incident, days int) bool {
	// get time at start of day at CreatedAt
	t := incident.CreatedAt
	targetTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	oneDay, _ := time.ParseDuration("24h")
	for days >= 0 {
		targetTime = targetTime.Add(oneDay)
		if isWeekDay(targetTime) {
			days--
		}
	}

	return targetTime.After(incident.SolvedAt)
}

func isWeekDay(moment time.Time) bool {
	return moment.Weekday() != time.Saturday && moment.Weekday() != time.Sunday
}

func getProdCategoriesToExclude() []string {
	filterCategories := []string{
		"HFC Network",
		"GIS Systems",
		"ACS/TR069 vDSL",
		"Field Force Management",
		"Monitoring",
		"Docsis",
		"Radio",
		"International Voice Unit",
		"Idefix",
		"Timecop",
		"Optical Transport",
	}
	return filterCategories
}

func filterOutProdCategories(incidents []Incident, categories []string) []Incident {
	var result []Incident
OuterLoop:
	for _, incident := range incidents {
		for _, category := range categories {
			if strings.Contains(incident.ProdCategory, category) {
				continue OuterLoop
			}
		}
		result = append(result, incident)
	}
	return result
}

func filterByCountry(incidents []Incident, country string) []Incident {
	var result []Incident
	for _, incident := range incidents {
		if strings.Contains(incident.Country, country) {
			result = append(result, incident)
		}
	}
	return result
}

func filterByMonthYear(incidents []Incident, month int, year int) []Incident {
	var result []Incident
	for _, incident := range incidents {
		if int(incident.CreatedAt.Month()) == month && incident.CreatedAt.Year() == year {
			result = append(result, incident)
		}
	}
	return result
}

func filterByPriority(incidents []Incident, priority int) []Incident {
	var result []Incident
	for _, incident := range incidents {
		if incident.Priority == priority {
			result = append(result, incident)
		}
	}
	return result
}

func getPreviousMonth(month int, year int) (int, int) {
	month--
	if month == 0 {
		month = 12
		year--
	}
	return month, year
}

func reportOnSixMonths(incidents []Incident, month int, year int, xls excelize.File) {
	percentStyle, _ := xls.NewStyle(`{"number_format": 9}`)
	var sixMonthIncidents []Incident
	// do it for 5 months
	for i := 6; i > 0; i-- {
		monthIncidents := filterByMonthYear(incidents, month, year)
		sixMonthIncidents = append(sixMonthIncidents, monthIncidents...)
		monthName := monthIncidents[1].CreatedAt.UTC().Format("Jan")
		axis, _ := excelize.CoordinatesToCellName(2+i, 3)
		_ = xls.SetCellStr("Overview", axis, monthName)
		axis, _ = excelize.CoordinatesToCellName(2+i, 10)
		_ = xls.SetCellStr("Overview", axis, monthName)
		axis, _ = excelize.CoordinatesToCellName(2+i, 17)
		_ = xls.SetCellStr("Overview", axis, monthName)
		for priorityIndex, priority := range []int{Critical, High, Medium, Low} {
			priorityIncidents := filterByPriority(monthIncidents, priority)
			total := 0
			slaMet := 0
			for _, incident := range priorityIncidents {
				if incident.SLAReady {
					total++
					if incident.SLAMet {
						slaMet++
					}
				}
			}
			percentage := float64(slaMet) / float64(total)
			axis, _ = excelize.CoordinatesToCellName(2+i, 4+priorityIndex)
			_ = xls.SetCellInt("Overview", axis, total)
			axis, _ = excelize.CoordinatesToCellName(2+i, 11+priorityIndex)
			_ = xls.SetCellInt("Overview", axis, slaMet)
			axis, _ = excelize.CoordinatesToCellName(2+i, 18+priorityIndex)
			_ = xls.SetCellFloat("Overview", axis, percentage, 2, 32)
			_ = xls.SetCellStyle("Overview", axis, axis, percentStyle)
		}
		month, year = getPreviousMonth(month, year)
	}
	addProdCategoriesToSheet(sixMonthIncidents, xls)
	addIncidentsToSheet(sixMonthIncidents, xls)
}

func setupExcelFile(xls excelize.File) {
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

func createCharts(xls excelize.File) {
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

func collectProdCategories(incidents []Incident) map[string]ProdCategory {
	prodCategories := make(map[string]ProdCategory)
	for _, incident := range incidents {
		category, found := prodCategories[incident.ProdCategory]
		if !found {
			category = ProdCategory{}
			prodCategories[incident.ProdCategory] = category
		}
		category.Total++
		if incident.SLAMet {
			category.SLAMet++
		}
		switch incident.Priority {
		case Critical:
			category.Critical++
		case High:
			category.High++
		case Medium:
			category.Medium++
		case Low:
			category.Low++
		}
		prodCategories[incident.ProdCategory] = category
	}
	return prodCategories
}

func addProdCategoriesToSheet(incidents []Incident, xls excelize.File) {
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

func addIncidentsToSheet(incidents []Incident, xls excelize.File) {
	xls.SetActiveSheet(xls.NewSheet("Incidents"))

	// setup the header row
	_ = xls.SetCellStr("Incidents", "A1", "ID")
	_ = xls.SetCellStr("Incidents", "B1", "Created")
	_ = xls.SetCellStr("Incidents", "C1", "Solved")
	_ = xls.SetCellStr("Incidents", "D1", "Time Open")
	_ = xls.SetCellStr("Incidents", "E1", "Priority")
	_ = xls.SetCellStr("Incidents", "F1", "Product Category")
	_ = xls.SetCellStr("Incidents", "G1", "Service CI")
	_ = xls.SetCellStr("Incidents", "H1", "SLA Met")
	_ = xls.SetCellStr("Incidents", "I1", "Description")

	maxProdLen := 1
	maxCILen := 1
	maxDescLen := 1

	for row, incident := range incidents {
		rowStr := strconv.Itoa(row + 2)

		_ = xls.SetCellValue("Incidents", "A"+rowStr, incident.ID)
		_ = xls.SetCellValue("Incidents", "B"+rowStr, incident.CreatedAt)
		if incident.SLAReady {
			_ = xls.SetCellValue("Incidents", "C"+rowStr, incident.SolvedAt)
		}
		_ = xls.SetCellValue("Incidents", "D"+rowStr, incident.OpenTime)
		_ = xls.SetCellValue("Incidents", "E"+rowStr, priorityNames[incident.Priority])
		_ = xls.SetCellValue("Incidents", "F"+rowStr, incident.ProdCategory)
		_ = xls.SetCellValue("Incidents", "G"+rowStr, incident.ServiceCI)
		_ = xls.SetCellValue("Incidents", "H"+rowStr, incident.SLAMet)
		_ = xls.SetCellValue("Incidents", "I"+rowStr, incident.Description)

		if len(incident.ProdCategory) > maxProdLen {
			maxProdLen = len(incident.ProdCategory)
		}
		if len(incident.ServiceCI) > maxCILen {
			maxCILen = len(incident.ServiceCI)
		}
		if len(incident.Description) > maxDescLen {
			maxDescLen = len(incident.Description)
		}
	}

	_ = xls.SetColWidth("Incidents", "A", "A", 16.0)
	_ = xls.SetColWidth("Incidents", "B", "B", 16.0)
	_ = xls.SetColWidth("Incidents", "C", "C", 16.0)
	_ = xls.SetColWidth("Incidents", "F", "F", 0.9*float64(maxProdLen))
	_ = xls.SetColWidth("Incidents", "G", "G", 0.9*float64(maxCILen))
	_ = xls.SetColWidth("Incidents", "I", "I", 0.9*float64(maxDescLen))

	rowStr := strconv.Itoa(len(incidents) + 1)
	_ = xls.AutoFilter("Incidents", "A1", "I"+rowStr, "")
}
