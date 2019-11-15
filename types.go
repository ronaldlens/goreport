package main

// Priority of incidents, int
const (
	Critical = iota
	High
	Medium
	Low
)

var priorityNames = []string{"Critical", "High", "Medium", "Low"}
var monthNames = []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

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

// ProdCategory is a struct describing a row in the product category tab
type ProdCategory struct {
	Total    int
	Critical int
	High     int
	Medium   int
	Low      int
	SLAMet   int
}
