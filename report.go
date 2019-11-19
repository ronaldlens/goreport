package main

import "log"

func runReport(incidents *Incidents, country string, month int, year int, splitArea bool, outputFilename string, minimumIncidents MinimumIncidents, verbose bool) {

	if outputFilename == "" {
		outputFilename = getFilename(country, month, year)
	}

	var sheet Sheet
	sheet.init()

	var totalIncidents Incidents
	if splitArea {
		itIncidents := incidents.filterByBusinessArea("IT")
		sheet.setupOverviewSheet("IT")
		itIncidents = itIncidents.reportOnSixMonths(month, year, "IT", &sheet, minimumIncidents)
		sheet.createCharts("IT")

		networkIncidents := incidents.filterByBusinessArea("Network")
		sheet.setupOverviewSheet("Network")
		networkIncidents = networkIncidents.reportOnSixMonths(month, year, "Network", &sheet, minimumIncidents)
		sheet.createCharts("Network")

		totalIncidents = append(itIncidents, networkIncidents...)

	} else {
		sheet.setupOverviewSheet("")
		totalIncidents = incidents.reportOnSixMonths(month, year, "", &sheet, minimumIncidents)
		sheet.createCharts("")
	}

	sheet.addProdCategoriesToSheet(totalIncidents)
	sheet.addIncidentsToSheet(totalIncidents)

	err := sheet.SaveAs(outputFilename)
	if err != nil {
		log.Fatalf("Error saving excel file: %v", err)
	}
	if verbose {
		log.Printf("Wrote output to %s", outputFilename)
	}

}
