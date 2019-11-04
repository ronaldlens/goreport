package main

import (
	"fmt"
	"time"
)

func checkSLA(incident *Incident) bool {
	// only process if the incident is solved
	if !incident.SLAReady {
		return false
	}
	switch incident.Priority {
	case Critical:
		return checkSLAHours(incident, 4)
	case High:
		return checkSLAHours(incident, 8)
	case Medium:
		return checkSLABusinessDays(incident, 1)
	case Low:
		return checkSLABusinessDays(incident, 3)
	}
	return false
}

func checkSLAHours(incident *Incident, hours int) bool {
	duration, _ := time.ParseDuration(fmt.Sprintf("%dh", hours))
	target := incident.CreatedAt.Add(duration)
	return target.After(incident.SolvedAt)
}

func checkSLABusinessDays(incident *Incident, days int) bool {
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

	return targetTime.After(incident.SolvedAt)
}

func isWeekDay(moment time.Time) bool {
	return moment.Weekday() != time.Saturday && moment.Weekday() != time.Sunday
}
