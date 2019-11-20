package main

import (
	"flag"
	"log"
	"strings"
	"time"
)

// command line arguments
var flagVars struct {
	configFilename    string
	inputFilename     string
	referenceFilename string
	outputFilename    string
	country           string
	month             int
	year              int
	now               bool
	verbose           bool
}

var config Config

func init() {
	// set up all command line flags
	flag.StringVar(&flagVars.configFilename, "cfg", "goreport.yaml", "Configuration filename")
	flag.StringVar(&flagVars.inputFilename, "input", "allincidents.csv", "Tab delimited incident input filename")
	flag.StringVar(&flagVars.referenceFilename, "reference", "", "Excel file to use as input reference")
	flag.StringVar(&flagVars.outputFilename, "output", "", "Output filename to use for xlsx file")
	flag.StringVar(&flagVars.country, "country", "", "Country to report on")

	flag.IntVar(&flagVars.month, "month", -1, "Month to report on (1..12)")
	flag.IntVar(&flagVars.year, "year", -1, "Year to report on")
	flag.BoolVar(&flagVars.now, "now", false, "Use current month instead of last month")

	flag.BoolVar(&flagVars.verbose, "v", false, "Increased verbosity")

}

func processCommandLineArgs() {
	flag.Parse()

	// read configuration file
	var err error
	config, err = readConfig(flagVars.configFilename)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// get country from config file
	// or defined via command lne args
	if flagVars.country == "" {
		flagVars.country = config.DefaultCountry
	}

	// if no month or year was supplied
	// use last month unless now was supplied as an option
	if flagVars.month == -1 || flagVars.year == -1 {
		flagVars.month = int(time.Now().Month())
		flagVars.year = time.Now().Year()
		if !flagVars.now {
			flagVars.month, flagVars.year = getPreviousMonth(flagVars.month, flagVars.year)
		}
	}
}

func processCommandLineCommand(incidents Incidents) {
	// work through commands
	if hasCommand("list") {
		runListCommand(incidents)
	} else if hasCommand("report") {
		runReportCommand(incidents)
	} else if hasCommand("gui") {
		RunGui(incidents)
	} else {
		log.Fatalf("No command specified")
	}
}

func runListCommand(incidents Incidents) {
	if hasNoun("countries") {
		if flagVars.verbose {
			log.Printf("Listing all countries")
		}
		listCountries(incidents)
	} else if hasNoun("prodcategories") {
		if flagVars.country != "" {
			if flagVars.verbose {
				log.Printf("Filtering by country %s", flagVars.country)
			}
			incidents = incidents.filterByCountry(flagVars.country)
		}
		listProductCategories(incidents)
	} else if hasNoun("services") {
		if flagVars.country != "" {
			if flagVars.verbose {
				log.Printf("Filtering by country %s", flagVars.country)
			}
			incidents = incidents.filterByCountry(flagVars.country)
		}
		listServices(incidents)
	}
}

func runReportCommand(incidents Incidents) {
	// reduce incidents
	countryConfig := getCountryFromConfig(config, flagVars.country)
	incidents = incidents.filterByCountry(flagVars.country)
	incidents = incidents.filterOutProdCategories(countryConfig.FilterOutCategories)

	// if we are to use a reference xlsx, process it
	// if the name equals to 'same' use the same name as the output
	// if the name equals to 'previous' or 'prev' use the xlsx from last month
	if flagVars.referenceFilename != "" {
		if flagVars.referenceFilename == "same" {
			flagVars.referenceFilename = getFilename(flagVars.country, flagVars.month, flagVars.year)
		} else if flagVars.referenceFilename == "previous" || flagVars.referenceFilename == "prev" {
			prevMonth, prevYear := getPreviousMonth(flagVars.month, flagVars.year)
			flagVars.referenceFilename = getFilename(flagVars.country, prevMonth, prevYear)
		}
		incidents = ProcessReferenceFile(incidents, flagVars.referenceFilename)
	}

	slaSet := ParseSLAConfig(countryConfig.SLAs)
	incidents = checkIncidentsAgainstSLA(incidents, slaSet)
	runReport(&incidents, flagVars.country, flagVars.month, flagVars.year, countryConfig.SplitArea,
		flagVars.outputFilename, countryConfig.MinimumIncidents, flagVars.verbose)
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
