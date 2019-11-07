package main

import (
	"flag"
	"fmt"
	"strings"

	"log"
	"time"
)

func main() {
	start := time.Now()

	// set up command line arguments
	var configFilename string
	var inputFilename string
	var outputFilename string
	var country string
	var month int
	var year int
	var verbose bool

	flag.StringVar(&configFilename, "cfg", "goreport.yaml", "Configuration filename")
	flag.StringVar(&inputFilename, "input", "allincidents.csv", "Tab delimited incident input filename")
	flag.StringVar(&outputFilename, "output", "", "Output filename to use for xlsx file")
	flag.StringVar(&country, "country", "", "Country to report on")

	flag.IntVar(&month, "month", -1, "Month to report on (1..12)")
	flag.IntVar(&year, "year", -1, "Year to report on")

	flag.BoolVar(&verbose, "v", false, "Increased verbosity")

	flag.Parse()

	config, err := readConfig(configFilename)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// process country
	if country == "" {
		country = config.DefaultCountry
	}

	incidents, err := importIncidents(inputFilename)
	if err != nil {
		log.Fatalf("Error importing %s: %v", inputFilename, err)
	}

	if verbose {
		log.Printf("Loaded a total of %d incidents from %s\n", len(incidents), inputFilename)
	}

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

	if outputFilename == "" {
		outputFilename = fmt.Sprintf("report-%s-%02d-%d.xlsx", strings.ToLower(country), month, year)
	}
	reportOnSixMonths(incidents, month, year, &sheet)
	sheet.createCharts()

	err = sheet.SaveAs(outputFilename)
	if err != nil {
		log.Fatalf("Error saving excel file: %v", err)
	}
	if verbose {
		log.Printf("Wrote output to %s", outputFilename)
		log.Printf("Total running time: %s\n", time.Since(start))
	}
}

func runReport() {

}
