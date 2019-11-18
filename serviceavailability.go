package main

type ReportPeriod struct {
	startMonth int
	startYear  int
	endMonth   int
	endYear    int
}

func getMinutesInMonth(month int, year int) int {
	return 30 * 24 * 60
}

func calculateSA(incidents Incidents, services []string, period ReportPeriod) map[string]float64 {
	month := period.startMonth
	year := period.startYear
	while month != period.endMonth {
		
	}
	month, year = getNextMonth(month, year)
}
