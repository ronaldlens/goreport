package main

import (
	"testing"
	"time"
)

func Test_checkSLAHours(t *testing.T) {
	timeStart := time.Date(2019, 10, 6, 9, 0, 0, 0, time.UTC)
	oneHour, _ := time.ParseDuration("1h")
	time1h := timeStart.Add(oneHour)
	time2h := time1h.Add(oneHour)
	time4h := time2h.Add(oneHour).Add(oneHour)

	incident := Incident{CreatedAt: timeStart, SolvedAt: time1h}
	if !checkSLAHours(&incident, 2) {
		t.Errorf("checkSLAHours=1h, SLA=2h")
	}
	incident.SolvedAt = time4h
	if checkSLAHours(&incident, 2) {
		t.Errorf("checkSLAHours=4h, SLA=2h")
	}

}

func Test_checkSLABusinessDays(t *testing.T) {
	// starting on Sunday
	timeStart := time.Date(2019, 10, 6, 9, 0, 0, 0, time.UTC)
	twelveHour, _ := time.ParseDuration("12h")
	timePlus12 := timeStart.Add(twelveHour)
	timePlus24 := timePlus12.Add(twelveHour)
	timePlus36 := timePlus24.Add(twelveHour)
	timePlus48 := timePlus36.Add(twelveHour)

	// start on Sunday, solve on Sunday
	incident := Incident{CreatedAt: timeStart, SolvedAt: timePlus12}
	if !checkSLABusinessDays(&incident, 1) {
		t.Errorf("checkSLABusinessDays created:%v solved:%v, SLA=1d", incident.CreatedAt, incident.SolvedAt)
	}

	// start on Sunday, solve on Monday
	incident.SolvedAt = timePlus24
	if !checkSLABusinessDays(&incident, 1) {
		t.Errorf("checkSLABusinessDays created:%v solved:%v, SLA=1d", incident.CreatedAt, incident.SolvedAt)
	}

	// start on Sunday, solve on Tuesday
	incident.SolvedAt = timePlus48
	if checkSLABusinessDays(&incident, 1) {
		t.Errorf("checkSLABusinessDays created:%v solved:%v, SLA=1d", incident.CreatedAt, incident.SolvedAt)
	}

	// Monday
	timeStart = time.Date(2019, 10, 7, 9, 0, 0, 0, time.UTC)
	incident.CreatedAt = timeStart
	timePlus12 = timeStart.Add(twelveHour)
	timePlus24 = timePlus12.Add(twelveHour)
	timePlus36 = timePlus24.Add(twelveHour)
	timePlus48 = timePlus36.Add(twelveHour)
	incident.SolvedAt = timePlus12

	// start on Monday, solve on Monday
	if !checkSLABusinessDays(&incident, 1) {
		t.Errorf("checkSLABusinessDays created:%v solved:%v, SLA=1d", incident.CreatedAt, incident.SolvedAt)
	}

	// start on Monday, solve on Tuesday
	incident.SolvedAt = timePlus24
	if !checkSLABusinessDays(&incident, 1) {
		t.Errorf("checkSLABusinessDays created:%v solved:%v, SLA=1d", incident.CreatedAt, incident.SolvedAt)
	}
}

func Test_isWeekDay(t *testing.T) {
	// Sat
	timestamp := time.Date(2019, 10, 5, 9, 0, 0, 0, time.UTC)
	if isWeekDay(timestamp) {
		t.Errorf("isWeekday returns true on Saturday %v", timestamp)
	}

	// Sun
	timestamp = time.Date(2019, 10, 6, 9, 0, 0, 0, time.UTC)
	if isWeekDay(timestamp) {
		t.Errorf("isWeekday returns true on Sunday %v", timestamp)
	}

	// Mon
	timestamp = time.Date(2019, 10, 7, 9, 0, 0, 0, time.UTC)
	if !isWeekDay(timestamp) {
		t.Errorf("isWeekday returns true on Monday %v", timestamp)
	}

	// Fri
	timestamp = time.Date(2019, 10, 4, 9, 0, 0, 0, time.UTC)
	if !isWeekDay(timestamp) {
		t.Errorf("isWeekday returns true on Friday %v", timestamp)
	}

}
