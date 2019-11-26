package main

import (
	"github.com/360EntSecGroup-Skylar/excelize"
	"strings"
	"time"
)

// Incidents type contains and array of Incident
type Incidents []Incident

// Incident struct described an incident
type Incident struct {
	Country           string
	ID                string
	CreatedAt         time.Time
	SolvedAt          time.Time
	Priority          int
	Description       string
	Resolution        string
	Service           string
	ProdCategory      string
	ServiceCI         string
	BusinessArea      string
	SLAReady          bool
	SLAMet            bool
	OpenTime          int
	CorrectedTime     string        // corrected time in string format
	CorrectedSolved   time.Time     // the new corrected solved time
	CorrectedOpenTime time.Duration // the open time as duration
	Exclude           bool
}

func (incidents *Incidents) filterOutProdCategories(categories []string) Incidents {
	var result []Incident

	// return unfiltered list if the argument nofilter was supplied
	if flagVars.nofilter {
		return *incidents
	}

	categoryMap := make(map[string]int)
	for _, category := range categories {
		categoryMap[category] = 1
	}

	for _, incident := range *incidents {
		if flagVars.reverse {
			if categoryMap[incident.ProdCategory] == 1 {
				result = append(result, incident)
			}
		} else {
			if categoryMap[incident.ProdCategory] != 1 {
				result = append(result, incident)
			}
		}
	}
	return result
}

func (incidents *Incidents) filterByCountry(country string) Incidents {
	var result []Incident
	for _, incident := range *incidents {
		if strings.Contains(incident.Country, country) {
			result = append(result, incident)
		}
	}
	return result
}

func (incidents *Incidents) filterByMonthYear(month int, year int) Incidents {
	var result []Incident
	for _, incident := range *incidents {
		if int(incident.CreatedAt.Month()) == month && incident.CreatedAt.Year() == year {
			result = append(result, incident)
		}
	}
	return result
}

func (incidents *Incidents) filterByPriority(priority int) Incidents {
	var result []Incident
	for _, incident := range *incidents {
		if incident.Priority == priority {
			result = append(result, incident)
		}
	}
	return result
}

func (incidents *Incidents) filterByBusinessArea(area string) Incidents {
	var result []Incident
	for _, incident := range *incidents {
		if incident.BusinessArea == area {
			result = append(result, incident)
		}
	}
	return result
}

func (incidents *Incidents) filterByService(service string) Incidents {
	var result []Incident
	for _, incident := range *incidents {
		if strings.EqualFold(incident.Service, service) {
			result = append(result, incident)
		}
	}
	return result
}

func (incidents *Incidents) collectProdCategories() map[string]ProdCategory {
	prodCategories := make(map[string]ProdCategory)
	for _, incident := range *incidents {
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

func (incidents *Incidents) reportOnSixMonths(month int, year int, area string, sheet *Sheet, minimumIncidentsConfig MinimumIncidents) Incidents {
	xls := sheet.file
	if area != "" {
		area = " " + area
	}
	//percentStyle, _ := xls.NewStyle(`{"number_format": 9}`)
	percentStyle2, _ := xls.NewStyle(`{"number_format": 10}`)
	greenStyle, _ := xls.NewStyle(`{"fill":{"type":"pattern","color":["#00FF00"],"pattern":1},"number_format": 9, "alignment":{"horizontal":"center"}}`)
	redStyle, _ := xls.NewStyle(`{"fill":{"type":"pattern","color":["#FF0000"],"pattern":1},"number_format": 9,"alignment":{"horizontal":"center"},"font":{"color":"#FFFFFF"}}`)
	greenStyle2, _ := xls.NewStyle(`{"fill":{"type":"pattern","color":["#00FF00"],"pattern":1},"number_format": 10, "alignment":{"horizontal":"center"}}`)
	redStyle2, _ := xls.NewStyle(`{"fill":{"type":"pattern","color":["#FF0000"],"pattern":1},"number_format": 10,"alignment":{"horizontal":"center"},"font":{"color":"#FFFFFF"}}`)

	// to collect incidents for 'Incidents' tab, contains all incidents for 6 months
	var sixMonthIncidents Incidents

	var totalIncidents [7][4]int
	var slaMetIncidents [7][4]int
	var calcTotalIncidents [7][4]int
	var calcSLAMetIncidents [7][4]int

	// start 6 months ago
	month, year = subtractMonths(month, year, 5)

	// repeat for 6 months
	for index := 0; index < 6; index++ {

		// set month to 0
		for i := 0; i < 4; i++ {
			totalIncidents[index][i] = 0
			slaMetIncidents[index][i] = 0
		}

		// get incidents for a month
		// add them to the grand list
		monthIncidents := incidents.filterByMonthYear(month, year)
		sixMonthIncidents = append(sixMonthIncidents, monthIncidents...)

		// go through all priorities
		// iterate over all incidents for that priority
		// and update the 2 counters
		for _, priority := range []int{Critical, High, Medium, Low} {
			priorityIncidents := monthIncidents.filterByPriority(priority)
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
			calcSLAMetIncidents[index][priority] = slaMetIncidents[index][priority]
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
				calcSLAMetIncidents[index+1][priority] += calcSLAMetIncidents[index][priority]
				calcSLAMetIncidents[index][priority] = 0
			}
		}
	}

	// rewind time by 6 months and iterate over the 6 months
	month, year = subtractMonths(month, year, 6)

	for index := 0; index < 6; index++ {

		// add the month label
		monthName := MonthNames[month]
		axis, _ := excelize.CoordinatesToCellName(3+index, 3)
		_ = xls.SetCellStr("Overview"+area, axis, monthName)
		axis, _ = excelize.CoordinatesToCellName(3+index, 10)
		_ = xls.SetCellStr("Overview"+area, axis, monthName)
		axis, _ = excelize.CoordinatesToCellName(3+index, 17)
		_ = xls.SetCellStr("Overview"+area, axis, monthName)

		for _, priority := range []int{Critical, High, Medium, Low} {
			axis, _ = excelize.CoordinatesToCellName(3+index, 4+priority)
			_ = xls.SetCellInt("Overview"+area, axis, totalIncidents[index][priority])
			axis, _ = excelize.CoordinatesToCellName(3+index, 11+priority)
			_ = xls.SetCellInt("Overview"+area, axis, slaMetIncidents[index][priority])

			if calcTotalIncidents[index][priority] != 0 {
				percentage := float64(calcSLAMetIncidents[index][priority]) / float64(calcTotalIncidents[index][priority])
				axis, _ = excelize.CoordinatesToCellName(3+index, 18+priority)
				_ = xls.SetCellFloat("Overview"+area, axis, percentage, 3, 64)
				if percentage < 0.8 {
					_ = xls.SetCellStyle("Overview"+area, axis, axis, redStyle)
				} else {
					_ = xls.SetCellStyle("Overview"+area, axis, axis, greenStyle)
				}
			}
		}
		month, year = getNextMonth(month, year)
	}

	// Service availability
	if area == " IT" || area == "" {

		// define the period, starting 6 months back from the reporting month
		startMonth, startYear := subtractMonths(month, year, 6)
		period := ReportPeriod{
			startMonth: startMonth,
			startYear:  startYear,
			endMonth:   month,
			endYear:    year,
		}

		// calculate the availability, use all incidents to allow going back one more month
		itAvailability := calculateSA(*incidents, ITServicesNames, period)

		// used to get the name of the month
		monthIdx := startMonth

		// set up the table
		axis, _ := excelize.CoordinatesToCellName(1, 23)
		_ = xls.SetCellStr("Overview"+area, axis, "IT Service Availability")
		axis, _ = excelize.CoordinatesToCellName(2, 24)
		_ = xls.SetCellStr("Overview"+area, axis, "Target")
		_ = xls.SetColWidth("Overview"+area, "A", "A", 16.22)

		// loop through all services to set the name and the target percentage
		for idx, service := range ITServicesNames {
			axis, _ := excelize.CoordinatesToCellName(1, 25+idx)
			_ = xls.SetCellStr("Overview"+area, axis, service)
			axis, _ = excelize.CoordinatesToCellName(2, 25+idx)
			_ = xls.SetCellFloat("Overview"+area, axis, 0.995, 3, 64)
			_ = xls.SetCellStyle("Overview"+area, axis, axis, percentStyle2)
		}

		// loop through the 6 months of the report
		for idx := 0; idx < 6; idx++ {
			axis, _ := excelize.CoordinatesToCellName(3+idx, 24)
			_ = xls.SetCellStr("Overview"+area, axis, MonthNames[monthIdx])
			for serviceIdx, service := range ITServicesNames {
				value := itAvailability[service][idx]
				axis, _ := excelize.CoordinatesToCellName(3+idx, 25+serviceIdx)
				_ = xls.SetCellFloat("Overview"+area, axis, value, 3, 64)
				_ = xls.SetCellStyle("Overview"+area, axis, axis, percentStyle2)
				if value < 0.995 {
					_ = xls.SetCellStyle("Overview"+area, axis, axis, redStyle2)
				} else {
					_ = xls.SetCellStyle("Overview"+area, axis, axis, greenStyle2)
				}
			}

			// increment month and check for year end rollover
			monthIdx++
			if monthIdx == 13 {
				monthIdx = 1
			}
		}
	}

	return sixMonthIncidents
}

// check if an incident is created in the previous month
func (incident *Incident) isCreatedInPrevMonthYear(month int, year int) bool {
	month, year = getPreviousMonth(month, year)
	return incident.CreatedAt.Year() == year && int(incident.CreatedAt.Month()) == month
}

// check if an incident is resolved in a specific month
func (incident *Incident) isResolvedInMonthYear(month int, year int) bool {
	var resolvedTime time.Time
	if incident.CorrectedTime != "" {
		resolvedTime = incident.CorrectedSolved
	} else {
		resolvedTime = incident.SolvedAt
	}
	return resolvedTime.Year() == year && int(resolvedTime.Month()) == month
}

func (incident *Incident) getOutageMinutesInMonth(month int, year int) int {
	if !incident.isResolvedInMonthYear(month, year) {
		return 0
	}

	// if incident is created in the previous month we need to do some math
	if incident.isCreatedInPrevMonthYear(month, year) {

	} else {

	}

	//TODO: implement getOutageMinutesInMonth
	return 1

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
