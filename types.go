package main

import "time"

const (
	Critical = iota
	High
	Medium
	Low
)

var priorityNames = []string{"Critical", "High", "Medium", "Low"}

type Incident struct {
	Country         string
	ID              string
	CreatedAt       time.Time
	SolvedAt        time.Time
	Priority        int
	Description     string
	Resolution      string
	Service         string
	ProdCategory    string
	ServiceCI       string
	BusinessArea    string
	SLAReady        bool
	SLAMet          bool
	OpenTime        int
	CorrectedTime   string
	CorrectedSolved time.Time
	Exclude         bool
}

type ProdCategory struct {
	Total    int
	Critical int
	High     int
	Medium   int
	Low      int
	SLAMet   int
}
