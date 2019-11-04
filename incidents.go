package main

import "strings"

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
