package main

import "time"

// ServiceAvailability is a map of float arrays
// the key (string) contains the service and the list the vluaes for the months
type ServiceAvailability map[string][]float64

// ReportPeriod defines the start and end of a reporting period
type ReportPeriod struct {
	startMonth int
	startYear  int
	endMonth   int
	endYear    int
}

// getMinutesInMonth returns the number of minutes in a given month
// time.Date with day 0 returns the day before the start of the month
// this gives number of days in the month
// returns the result in a ServiceAvailability map
func getMinutesInMonth(month int, year int) int {
	month, year = getNextMonth(month, year)
	return time.Date(year, time.Month(month), 0, 0, 0, 0, 0, time.UTC).Day() * 24 * 60
}

func calculateSA(incidents Incidents, services []string, period ReportPeriod) ServiceAvailability {
	// start at the beginning
	month := period.startMonth
	year := period.startYear

	result := make(ServiceAvailability)

	// go through the months until the endMonth & Year is reached
	for {
		totMinutes := getMinutesInMonth(month, year)

		monthIncidents := incidents.filterByMonthYear(month, year)
		outageMinutes := 0

		for _, service := range services {
			// go through incidents for a service in this month to get the outage minutes
			serviceIncidents := monthIncidents.filterByService(service)
			for _, incident := range serviceIncidents {
				if incident.SLAReady {
					outageMinutes += incident.OpenTime
				}
			}

			availability := float64(totMinutes-outageMinutes) / float64(totMinutes)

			result[service] = append(result[service], availability)
		}

		if month == period.endMonth && year == period.endYear {
			break
		}
		month, year = getNextMonth(month, year)
	}

	return result
}
