package main

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

// SLA struct in the configuration file
// Either Hours or Days has a value
type SLA struct {
	Priority string
	Hours    int
	Days     int
}

// MinimumIncidents used for TTR performance measurement
// If minimum not reached, value will carry over to the next month
type MinimumIncidents struct {
	Critical int
	High     int
	Medium   int
	Low      int
}

// Country struct holds the configuration for a given country
type Country struct {
	Name                 string
	SplitArea            bool
	ITServiceWindow      string
	ITServiceWindowStart time.Time
	ITServiceWindowEnd   time.Time
	SLAs                 []SLA
	MinimumIncidents     MinimumIncidents
	FilterOutCategories  []string
}

// Config struct contains the overall configuration
// Default country is optional
type Config struct {
	DefaultCountry  string
	OutputDirectory string
	Countries       []Country
}

func readConfig(filename string) (Config, error) {
	if flagVars.verbose {
		log.Printf("Loading config file: %s", filename)
	}
	config := Config{}
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func getCountryFromConfig(config Config, countryName string) Country {
	for _, country := range config.Countries {
		if country.Name == countryName {
			return country
		}
	}
	log.Fatalf("Cannot find country %s in configuration", countryName)
	return Country{}
}
