package main

import (
	"flag"
	"strings"

	"log"
	"time"
)

// command line arguments
var configFilename string
var inputFilename string
var referenceFilename string
var outputFilename string
var country string
var month int
var year int
var now bool
var verbose bool

func init() {
	// set up all command line flags
	flag.StringVar(&configFilename, "cfg", "goreport.yaml", "Configuration filename")
	flag.StringVar(&inputFilename, "input", "allincidents.csv", "Tab delimited incident input filename")
	flag.StringVar(&referenceFilename, "reference", "", "Excel file to use as input reference")
	flag.StringVar(&outputFilename, "output", "", "Output filename to use for xlsx file")
	flag.StringVar(&country, "country", "", "Country to report on")

	flag.IntVar(&month, "month", -1, "Month to report on (1..12)")
	flag.IntVar(&year, "year", -1, "Year to report on")
	flag.BoolVar(&now, "now", false, "Use current month instead of last month")

	flag.BoolVar(&verbose, "v", false, "Increased verbosity")

}

func main() {
	// start time measurement and parse command line flags
	start := time.Now()
	flag.Parse()

	// read configuration file
	config, err := readConfig(configFilename)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// get country from config file
	// or defined via command lne args
	if country == "" {
		country = config.DefaultCountry
	}

	// if no month or year was supplied
	// use last month unless now was supplied as an option
	if month == -1 || year == -1 {
		month = int(time.Now().Month())
		year = time.Now().Year()
		if !now {
			month, year = getPreviousMonth(month, year)
		}
	}

	// load the incidents
	incidents, err := importIncidents(inputFilename)
	if err != nil {
		log.Fatalf("Error importing %s: %v", inputFilename, err)
	}
	if verbose {
		log.Printf("Loaded a total of %d incidents from %s\n", len(incidents), inputFilename)
	}

	// work through commands
	if hasCommand("list") {
		if hasNoun("countries") {
			if verbose {
				log.Printf("Listing all countries")
			}
			listCountries(incidents)
		}
		if hasNoun("prodcategories") {
			if country != "" {
				if verbose {
					log.Printf("Filtering by country %s", country)
				}
				incidents = incidents.filterByCountry(country)
			}
			listProductCategories(incidents)
		}
	} else if hasCommand("report") {
		// reduce incidents
		countryConfig := getCountryFromConfig(config, country)
		incidents = incidents.filterByCountry(country)
		incidents = incidents.filterOutProdCategories(countryConfig.FilterOutCategories)

		// if we are to use a reference xlsx, process it
		// if the name equals to 'same' use the same name as the output
		// if the name equals to 'previous' or 'prev' use the xlsx from last month
		if referenceFilename != "" {
			if referenceFilename == "same" {
				referenceFilename = getFilename(country, month, year)
			} else if referenceFilename == "previous" || referenceFilename == "prev" {
				prevMonth, prevYear := getPreviousMonth(month, year)
				referenceFilename = getFilename(country, prevMonth, prevYear)
			}
			incidents = ProcessReferenceFile(incidents, referenceFilename)
		}

		slaSet := ParseSLAConfig(countryConfig.SLAs)
		incidents = checkIncidentsAgainstSla(incidents, slaSet)
		runReport(&incidents, country, month, year, outputFilename, countryConfig.MinimumIncidents, verbose)
	} else {
		log.Fatalf("No command specified")
	}

	if verbose {
		log.Printf("Total running time: %s\n", time.Since(start))
	}
}

func runReport(incidents *Incidents, country string, month int, year int, outputFilename string, minimumIncidents MinimumIncidents, verbose bool) {

	if outputFilename == "" {
		outputFilename = getFilename(country, month, year)
	}

	var sheet Sheet
	sheet.init()
	sheet.setupExcelFile("IT")

	incidents.reportOnSixMonths(month, year, "IT", &sheet, minimumIncidents)
	sheet.createCharts("IT")

	err := sheet.SaveAs(outputFilename)
	if err != nil {
		log.Fatalf("Error saving excel file: %v", err)
	}
	if verbose {
		log.Printf("Wrote output to %s", outputFilename)
	}

}

// check if the command line contains a specific command (verb)
func hasCommand(command string) bool {
	// check is enough args
	if len(flag.Args()) < 1 {
		return false
	}

	// check if arg is an option
	if strings.HasPrefix(flag.Args()[0], "-") {
		return false
	}

	// check if command is given
	return flag.Args()[0] == command
}

// used after the command is recognized, check for a noun
// example ./goreport list countries
// list is command, countries is noun
func hasNoun(noun string) bool {
	// check is enough args
	if len(flag.Args()) < 2 {
		return false
	}

	// check if arg is an option
	if strings.HasPrefix(flag.Args()[1], "-") {
		return false
	}

	// check if command is given
	return flag.Args()[1] == noun

}
