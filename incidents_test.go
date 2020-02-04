package main

import (
	"testing"
	"time"
)

func TestIncidents_filterOutProdCategories(t *testing.T) {
	i1 := Incident{ID: "1", ProdCategory2: "foo"}
	i2 := Incident{ID: "2", ProdCategory2: "foo"}
	i3 := Incident{ID: "3", ProdCategory2: "bar"}

	incidents := Incidents{i1, i2, i3}
	prodCategories := []string{"bar"}

	filtered := incidents.filterOutProdCategories(prodCategories)
	if len(filtered) != 2 {
		t.Errorf("Exepcted length of 2, got %d", len(filtered))
	}
}

func TestIncidents_filterByCountry(t *testing.T) {
	i1 := Incident{ID: "1", Country: "foo"}
	i2 := Incident{ID: "2", Country: "bar"}
	i3 := Incident{ID: "3", Country: "foo"}

	incidents := Incidents{i1, i2, i3}
	filtered := incidents.filterByCountry("foo")
	if len(filtered) != 2 {
		t.Errorf("Exepcted length of 2, got %d", len(filtered))
	}
}

func TestIncidents_filterByPriority(t *testing.T) {
	i1 := Incident{ID: "1", Priority: Critical}
	i2 := Incident{ID: "2", Priority: High}
	i3 := Incident{ID: "3", Priority: Critical}

	incidents := Incidents{i1, i2, i3}
	filtered := incidents.filterByPriority(Critical)
	if len(filtered) != 2 {
		t.Errorf("Exepcted length of 2, got %d", len(filtered))
	}
}

func Test_subtractMonths(t *testing.T) {
	type args struct {
		month int
		year  int
		delta int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
	}{
		{
			name: "simple date test",
			args: args{
				month: 10,
				year:  2019,
				delta: 4,
			},
			want:  6,
			want1: 2019,
		},
		{
			name: "simple date test across year",
			args: args{
				month: 1,
				year:  2019,
				delta: 4,
			},
			want:  9,
			want1: 2018,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := subtractMonths(tt.args.month, tt.args.year, tt.args.delta)
			if got != tt.want {
				t.Errorf("subtractMonths() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("subtractMonths() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getNextMonth(t *testing.T) {
	type args struct {
		month int
		year  int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
	}{
		{
			name: "simple 1",
			args: args{
				month: 10,
				year:  2019,
			},
			want:  11,
			want1: 2019,
		},
		{
			name: "simple 1 across year",
			args: args{
				month: 12,
				year:  2019,
			},
			want:  1,
			want1: 2020,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getNextMonth(tt.args.month, tt.args.year)
			if got != tt.want {
				t.Errorf("getNextMonth() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getNextMonth() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getPreviousMonth(t *testing.T) {
	type args struct {
		month int
		year  int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
	}{
		{
			name: "simple 1",
			args: args{
				month: 10,
				year:  2019,
			},
			want:  9,
			want1: 2019,
		},
		{
			name: "simple 1 across year",
			args: args{
				month: 1,
				year:  2019,
			},
			want:  12,
			want1: 2018,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getPreviousMonth(tt.args.month, tt.args.year)
			if got != tt.want {
				t.Errorf("getPreviousMonth() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getPreviousMonth() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestIncident_getOutageMinutesInMonth(t *testing.T) {
	startTime := time.Date(2019, time.April, 1, 0, 0, 0, 0, time.UTC)
	resolvedTime := time.Date(2019, time.April, 1, 1, 30, 0, 0, time.UTC)

	incident := Incident{CreatedAt: startTime, SolvedAt: resolvedTime, OpenTime: 90, SLAReady: true}

	got := incident.getOutageMinutesInMonth(4, 2019)
	if got != 90 {
		t.Errorf("getOutageMinutesInMonth got %d, want %d", got, 90)
	}

	// incident starts in the month before
	startTime = time.Date(2019, time.March, 31, 23, 0, 0, 0, time.UTC)
	incident = Incident{CreatedAt: startTime, SolvedAt: resolvedTime, OpenTime: 90, SLAReady: true}
	got = incident.getOutageMinutesInMonth(4, 2019)
	if got != 90 {
		t.Errorf("getOutageMinutesInMonth got %d, want %d", got, 90)
	}

}
