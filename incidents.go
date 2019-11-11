package main

import (
	"github.com/360EntSecGroup-Skylar/excelize"
	"strings"
)

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
			if total != 0 {
				percentage := float64(slaMet) / float64(total)
				axis, _ = excelize.CoordinatesToCellName(2+i, 18+priorityIndex)
				_ = xls.SetCellFloat("Overview", axis, percentage, 2, 32)
				_ = xls.SetCellStyle("Overview", axis, axis, percentStyle)

			}
			axis, _ = excelize.CoordinatesToCellName(2+i, 4+priorityIndex)
			_ = xls.SetCellInt("Overview", axis, total)
			axis, _ = excelize.CoordinatesToCellName(2+i, 11+priorityIndex)
			_ = xls.SetCellInt("Overview", axis, slaMet)
		}
		month, year = getPreviousMonth(month, year)
	}
	sheet.addProdCategoriesToSheet(sixMonthIncidents)
	sheet.addIncidentsToSheet(sixMonthIncidents)
}

func getPreviousMonth(month int, year int) (int, int) {
	month--
	if month == 0 {
		month = 12
		year--
	}
	return month, year
}
