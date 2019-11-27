package main

import "fmt"

func listCountries(incidents []Incident) {
	countries := make(map[string]int)
	for _, incident := range incidents {
		countries[incident.Country] = 1
	}

	for country := range countries {
		fmt.Println(country)
	}

}

func listProductCategories(incidents []Incident) {
	categories := make(map[string]int)
	for _, incident := range incidents {
		categories[incident.ProdCategory2] = 1
	}

	for category := range categories {
		fmt.Println(category)
	}
}

func listServices(incidents []Incident) {
	services := make(map[string]int)
	for _, incident := range incidents {
		services[incident.Service] = 1
	}

	for category := range services {
		fmt.Println(category)
	}
}
