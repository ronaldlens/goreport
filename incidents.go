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

func reportOnSixMonths(incidents []Incident, month int, year int, sheet *Sheet, minimumIncidentsConfig MinimumIncidents) {
	xls := sheet.file
	percentStyle, _ := xls.NewStyle(`{"number_format": 9}`)

	// to collect incidents for 'Incidents' tab, contains all incidents for 6 months
	var sixMonthIncidents []Incident

	var totalIncidents [7][4]int
	var slaMetIncidents [7][4]int
	var calcTotalIncidents [7][4]int
	var calcSlaMetIncidents [7][4]int

	// start 6 months ago
	month, year = subtractMonths(month, year, 6)

	// repeat for 6 months
	for index := 0; index < 6; index++ {

		// set month to 0
		for i := 0; i < 4; i++ {
			totalIncidents[index][i] = 0
			slaMetIncidents[index][i] = 0
		}

		// get incidents for a month
		// add them to the grand list
		monthIncidents := filterByMonthYear(incidents, month, year)
		sixMonthIncidents = append(sixMonthIncidents, monthIncidents...)

		// go through all priorities
		// iterate over all incidents for that priority
		// and update the 2 counters
		for _, priority := range []int{Critical, High, Medium, Low} {
			priorityIncidents := filterByPriority(monthIncidents, priority)
			for _, incident := range priorityIncidents {
				if incident.SLAReady {
					totalIncidents[index][priority]++
					if incident.SLAMet {
						slaMetIncidents[index][priority]++
					}
				}
			}

			// copy to the value used to calculate performance
			calcTotalIncidents[index][priority] = totalIncidents[index][priority]
			calcSlaMetIncidents[index][priority] = slaMetIncidents[index][priority]
		}

		// advance month, check for year rollover
		month, year = getNextMonth(month, year)
	}

	// process minimum incidents config
	var minimumIncidents [4]int
	minimumIncidents[0] = minimumIncidentsConfig.Critical
	minimumIncidents[1] = minimumIncidentsConfig.High
	minimumIncidents[2] = minimumIncidentsConfig.Medium
	minimumIncidents[3] = minimumIncidentsConfig.Low

	// run through the 6 months to check the minimum incident threshold
	// if the minimum is not reached, it moves forward until it is
	// or the end of the report is reached
	for index := 0; index < 6; index++ {
		for priority := Critical; priority <= Low; priority++ {
			if calcTotalIncidents[index][priority] < minimumIncidents[priority] {
				calcTotalIncidents[index+1][priority] += calcTotalIncidents[index][priority]
				calcTotalIncidents[index][priority] = 0
				calcSlaMetIncidents[index+1][priority] += calcSlaMetIncidents[index][priority]
				calcSlaMetIncidents[index][priority] = 0
			}
		}
	}

	// rewind time by 6 months and iterate over the 6 months
	month, year = subtractMonths(month, year, 6)

	for index := 0; index < 6; index++ {

		// add the month label
		monthName := monthNames[month]
		axis, _ := excelize.CoordinatesToCellName(3+index, 3)
		_ = xls.SetCellStr("Overview", axis, monthName)
		axis, _ = excelize.CoordinatesToCellName(3+index, 10)
		_ = xls.SetCellStr("Overview", axis, monthName)
		axis, _ = excelize.CoordinatesToCellName(3+index, 17)
		_ = xls.SetCellStr("Overview", axis, monthName)

		for _, priority := range []int{Critical, High, Medium, Low} {
			axis, _ = excelize.CoordinatesToCellName(3+index, 4+priority)
			_ = xls.SetCellInt("Overview", axis, totalIncidents[index][priority])
			axis, _ = excelize.CoordinatesToCellName(3+index, 11+priority)
			_ = xls.SetCellInt("Overview", axis, slaMetIncidents[index][priority])

			if calcTotalIncidents[index][priority] != 0 {
				percentage := float64(calcSlaMetIncidents[index][priority]) / float64(calcTotalIncidents[index][priority])
				axis, _ = excelize.CoordinatesToCellName(3+index, 18+priority)
				_ = xls.SetCellFloat("Overview", axis, percentage, 2, 32)
				_ = xls.SetCellStyle("Overview", axis, axis, percentStyle)
			}
		}
		month, year = getNextMonth(month, year)
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

func getNextMonth(month int, year int) (int, int) {
	month++
	if month == 13 {
		month = 1
		year++
	}
	return month, year
}

func subtractMonths(month int, year int, delta int) (int, int) {
	month -= delta
	if month < 1 {
		month += 12
		year--
	}
	return month, year
}
