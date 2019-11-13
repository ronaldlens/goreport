package main

const (
	Critical = iota
	High
	Medium
	Low
)

var priorityNames = []string{"Critical", "High", "Medium", "Low"}
var monthNames = []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
var ITServicesNames = []string{
	"CRM",
	"Billing",
	"Provisioning",
	"Web",
	"DTV",
	"Service Assurance",
	"ERP",
	"Middleware",
	"Infrastructure",
	"Hosted services"}

type ProdCategory struct {
	Total    int
	Critical int
	High     int
	Medium   int
	Low      int
	SLAMet   int
}
