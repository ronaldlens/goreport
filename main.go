package main

import (
	"flag"
	"github.com/spf13/viper"
	"log"
	"time"

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

func getPreviousMonth(month int, year int) (int, int) {
	month--
	if month == 0 {
		month = 12
		year--
	}
	return month, year
}
