package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type SLA struct {
	Priority string
	Hours    int
	Days     int
}

type MinimumIncidents struct {
	Critical int
	High     int
	Medium   int
	Low      int
}

type Country struct {
	Name                string
	SplitArea           bool
	SLAs                []SLA
	MinimumIncidents    MinimumIncidents
	FilterOutCategories []string
}

type Config struct {
	DefaultCountry string
	Countries      []Country
}

func readConfig(filename string) (Config, error) {
	config := Config{}
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return config, err
	}
	//fmt.Printf("read config:\n%v\n", config)
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
