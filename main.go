package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/spf13/pflag"
)

func main() {
	start := time.Now()

	// set up command line arguments
	flag.String("f", "allincidents.csv", "filename to load")
	flag.Int("m", -1, "month to retrieve")
	flag.Int("y", -1, "year to retrieve")
	flag.String("c", "Austria", "Country to retrieve")
	flag.String("x", "out.xlsx", "Excel file to write")
	flag.Bool("g", false, "use gui")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatalf("Error parsing commandline options: %v", err)
	}

	if viper.GetBool("g") {
		RunGui()
	} else {

		// get the command line arguments
		filename := viper.GetString("f")
		country := viper.GetString("c")
		month := viper.GetInt("m")
		year := viper.GetInt("y")
		xlsFilename := viper.GetString("x")

		incidents, err := importIncidents(filename)
		if err != nil {
			log.Fatalf("Error importing %s: %v", filename, err)
		}

		log.Printf("Loaded a total of %d incidents from %s\n", len(incidents), filename)

		// reduce incidents
		incidents = filterByCountry(incidents, country)
		incidents = filterOutProdCategories(incidents, getProdCategoriesToExclude())

		var sheet Sheet
		sheet.init()
		sheet.setupExcelFile()

		if month == -1 || year == -1 {
			month = int(time.Now().Month())
			year = time.Now().Year()
			month, year = getPreviousMonth(month, year)
		}

		reportOnSixMonths(incidents, month, year, &sheet)
		sheet.createCharts()

		err = sheet.SaveAs(xlsFilename)
		if err != nil {
			log.Fatalf("Error saving excel file: %v", err)
		}
		log.Printf("Wrote output to %s", xlsFilename)
	}
	log.Printf("Total running time: %s\n", time.Since(start))
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
		"Bumblebee",
		"Finanzarchiv AT",
		"HR",
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

func reportOnSixMonths(incidents []Incident, month int, year int, sheet *Sheet) {
	xls := sheet.file
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
	sheet.addProdCategoriesToSheet(sixMonthIncidents)
	sheet.addIncidentsToSheet(sixMonthIncidents)
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
