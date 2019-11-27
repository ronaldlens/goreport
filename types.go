package main

// Priority of incidents, int
const (
	Critical = iota
	High
	Medium
	Low
)

// PriorityNames is an array continaing strings describing the priority
var PriorityNames = []string{"Critical", "High", "Medium", "Low"}

// MonthNames is an array continaing the short names of months. It is 1 based, not 0 based!
var MonthNames = []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// ITServicesNames is used for determining service availability in IT
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

// NWServiceNames is used to determine the service availability on the network side
var NWServiceNames = []string{
	"Internet",
	"Voice",
	"CATV",
	"D4A",
	"Dawn",
	"Horizon TV",
	"Horizon GO",
}

// ProdCategory2 is a struct describing a row in the product category tab
type ProdCategory struct {
	Total    int
	Critical int
	High     int
	Medium   int
	Low      int
	SLAMet   int
}
