package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type SLA struct {
	Priority string
	Hours    int
	Days     int
}

type Country struct {
	Name                string
	SLAs                []SLA
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
