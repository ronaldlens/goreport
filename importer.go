package main

import (
	"bufio"
	"errors"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"log"
	"os"
	"strings"
	"time"
)

func importIncidents(filename string) ([]Incident, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Trying to open %s: %v", filename, err)
		return nil, err
	}

	scanner := bufio.NewScanner(transform.NewReader(
		file, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))

	// get header line and build map to reference column numbers by the name
	scanner.Scan()
	headerParts := strings.Split(scanner.Text(), "\t")
	headers, err := parseHeaders(headerParts)
	if err != nil {
		log.Printf("Error parsing header: %v", err)
		return nil, err
	}

	var incidents []Incident
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		var inc = Incident{
			Country:      parts[headers["Country"]],
			ID:           parts[4],
			Description:  parts[13],
			Resolution:   parts[14],
			ProdCategory: parts[8],
			Service:      parts[9],
			ServiceCI:    parts[10],
			BusinessArea: parts[11],
		}

		switch parts[headers["Priority"]] {
		case "Critical":
			inc.Priority = Critical
		case "High":
			inc.Priority = High
		case "Medium":
			inc.Priority = Medium
		case "Low":
			inc.Priority = Low
		}

		// get timestamps, createdAt and solvedAt
		// solvedAt may not be filled in (yet), if so, don't use it for SLA calculations
		t, err := parseTimeStamp(parts[5])
		if err == nil {
			inc.CreatedAt = t
		}
		t, err = parseTimeStamp(parts[6])
		if err == nil {
			inc.SolvedAt = t
			inc.SLAReady = true
		}

		// calculate the open time in minutes
		// check if the SLA is met
		if inc.SLAReady {
			inc.OpenTime = int(inc.SolvedAt.Sub(inc.CreatedAt).Minutes())
		} else {
			inc.OpenTime = 0
		}
		inc.SLAMet = checkSLA(&inc)

		incidents = append(incidents, inc)
	}

	return incidents, nil
}

func parseHeaders(headerParts []string) (map[string]int, error) {
	requiredFields := []string{
		"Country",
		"Incident Number",
		"Create DateTime",
		"Last Resolved DateTime",
		"Priority",
		"Product Categorization Tier2",
		"Service",
		"Service CI",
		"Business area",
		"Status",
		"Description",
		"Resolution Description",
	}

	headers := make(map[string]int)
	for index, header := range headerParts {
		headers[header] = index
	}

	for _, requiredField := range requiredFields {
		_, exists := headers[requiredField]
		if !exists {
			return nil, errors.New("cannot find field")
		}
	}

	return headers, nil
}

func parseTimeStamp(ts string) (time.Time, error) {
	ts = strings.Replace(ts, " ", "T", 1) // Make it RFC3339 compliant
	ts = strings.Replace(ts, "/", "-", 2) // fix the date part
	ts += "Z"                             // add the Z for UTC
	return time.Parse(time.RFC3339, ts)
}
