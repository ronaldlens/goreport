package main

import (
	"flag"
	"github.com/spf13/viper"
	"log"
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

func getProdCategoriesToExclude() []string {
	filterCategories := []string{
		"HFC Network",
		"GIS Systems",
		"ACS/TR069 vDSL",
		"Field Force Management",
		"Monitoring",
		"DOCSIS",
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
