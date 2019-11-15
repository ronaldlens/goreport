package main

import (
	"fmt"
	"log"
	"time"
)

// SLAEntry is a struct describing SLA for a given priority,
// Either hours or days has a value, the other defaults to 0.
// days means business days
type SLAEntry struct {
	hours int
	days  int
}

// StringToPriority converts a string describing the priority to int
func StringToPriority(str string) int {
	switch str {
	case "Critical":
		return Critical
	case "High":
		return High
	case "Medium":
		return Medium
	case "Low":
		return Low
	default:
		return Low
	}
}

// PriorityToString converts priority as an int to a string describing the name
func PriorityToString(priority int) string {
	return priorityNames[priority]
}

// ParseSLAConfig takes an array of SLA from the configuration and converts
// it to an array of SLAEntry structs
func ParseSLAConfig(slaConfig []SLA) [4]SLAEntry {
	slaSet := [4]SLAEntry{}
	for _, slaConfigEntry := range slaConfig {
		id := StringToPriority(slaConfigEntry.Priority)
		slaSet[id].days = slaConfigEntry.Days
		slaSet[id].hours = slaConfigEntry.Hours
	}
	return slaSet
}

func checkIncidentsAgainstSLA(incidents []Incident, slaSet [4]SLAEntry) []Incident {
	var slaIncidents []Incident
	for _, incident := range incidents {
		incident.SLAMet = checkSLA(incident, slaSet)
		slaIncidents = append(slaIncidents, incident)
	}
	return slaIncidents
}

func checkSLA(incident Incident, slaSet [4]SLAEntry) bool {

	// only process if the incident is solved
	if !incident.SLAReady {
		return false
	}
	if slaSet[incident.Priority].days == 0 {
		return checkSLAHours(incident, slaSet[incident.Priority].hours)
	}
	return checkSLABusinessDays(incident, slaSet[incident.Priority].days)

}

func checkSLAHours(incident Incident, hours int) bool {
	duration, _ := time.ParseDuration(fmt.Sprintf("%dh", hours))
	target := incident.CreatedAt.Add(duration)
	if incident.CorrectedTime != "" {
		correctedDuration, err := time.ParseDuration(incident.CorrectedTime)
		if err != nil {
			log.Fatalf("Error parsing corrected duration for incident %s: %v", incident.ID, err)
		}
		incident.CorrectedSolved = incident.CreatedAt.Add(correctedDuration)
		return target.After(incident.CorrectedSolved)
	}
	return target.After(incident.SolvedAt)
}

func checkSLABusinessDays(incident Incident, days int) bool {
	// get time at start of day at CreatedAt
	t := incident.CreatedAt
	targetTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	oneDay, _ := time.ParseDuration("24h")
	for days >= 0 {
		targetTime = targetTime.Add(oneDay)
		if isWeekDay(targetTime) {
			days--
		}
	}

	if incident.CorrectedTime != "" {
		correctedDuration, err := time.ParseDuration(incident.CorrectedTime)
		if err != nil {
			log.Fatalf("Error parsing corrected duration for incident %s: %v", incident.ID, err)
		}
		incident.CorrectedSolved = incident.CreatedAt.Add(correctedDuration)
		return targetTime.After(incident.CorrectedSolved)
	}
	return targetTime.After(incident.SolvedAt)
}

func isWeekDay(moment time.Time) bool {
	return moment.Weekday() != time.Saturday && moment.Weekday() != time.Sunday
}
