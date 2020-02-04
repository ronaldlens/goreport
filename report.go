package main

import (
	"log"
	"path/filepath"
)

func runReport(incidents *Incidents, localIncidents *Incidents, country string, month int, year int, splitArea bool,
	outputFilename string, minimumIncidents MinimumIncidents, verbose bool, outputDirectory string) {

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
	sheet.addIncidentsToSheet(totalIncidents, "Incidents")
	//sheet.addIncidentsToSheet(localIncidents, "Local Incidents")
	if outputDirectory != "" {
		outputFilename = filepath.Join(outputDirectory, outputFilename)

	}
	err := sheet.SaveAs(outputFilename)
	if err != nil {
		log.Fatalf("Error saving excel file: %v", err)
	}
	if verbose {
		log.Printf("Wrote output to %s", outputFilename)
	}

}
