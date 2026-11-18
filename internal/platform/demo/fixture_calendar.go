package demo

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// FamilyCalendarURL is the ICS subscription of the showcase's calendar.
const FamilyCalendarURL = "https://calendar.demo.invalid/family.ics"

// DHLAPIKey is the example key of the parcel tracking; DHLParcels are
// tracked live through the fixture below.
const DHLAPIKey = "demo-only-dhl-key" //nolint:gosec // example key, answered only by the fixture

// DHLParcels are the example shipments followed with DHL tracking.
var DHLParcels = []struct{ Number, Label string }{
	{"JJD000390007775123456", "New router"},
	{"JJD000390007775654321", "Books from the antiquarian bookshop"},
}

func init() {
	handle(hostIs("calendar.demo.invalid"), func(_ *http.Request, now time.Time) (int, string, string) {
		return http.StatusOK, "text/calendar", familyCalendar(now)
	})
	handle(hostIs("api-eu.dhl.com"), dhlTracking)
}

// familyCalendar is an ICS feed with appointments across the coming two
// weeks, relative to today so the widget is never empty.
func familyCalendar(now time.Time) string {
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	events := []struct {
		days, hour, minute, minutes int
		title, location             string
		allDay                      bool
	}{
		{0, 18, 30, 90, "Football training", "Sportpark Donnerschwee", false},
		{1, 9, 0, 45, "Team call", "Online", false},
		{1, 16, 15, 30, "Dentist", "Praxis am Schlossplatz", false},
		{2, 19, 30, 150, "Birthday dinner for Anna", "Restaurant am Hafen", false},
		{3, 0, 0, 0, "Paper waste collection", "", true},
		{4, 10, 0, 120, "Farmers' market", "Rathausmarkt", false},
		{6, 15, 0, 60, "Parents' evening", "Grundschule Bürgerfelde", false},
		{8, 18, 30, 90, "Football training", "Sportpark Donnerschwee", false},
		{9, 7, 45, 0, "Train to Hamburg", "Oldenburg Hbf", false},
		{11, 0, 0, 0, "Grandma's birthday", "", true},
		{13, 20, 0, 120, "Concert: Nordlicht Ensemble", "Kulturetage", false},
	}
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Waylight demo//family//EN\r\nX-WR-CALNAME:Family\r\n")
	stamp := now.UTC().Format("20060102T150405Z")
	for i, e := range events {
		fmt.Fprintf(&b, "BEGIN:VEVENT\r\nUID:demo-family-%d@demo.invalid\r\nDTSTAMP:%s\r\nSUMMARY:%s\r\n", i, stamp, e.title)
		start := day.AddDate(0, 0, e.days)
		if e.allDay {
			fmt.Fprintf(&b, "DTSTART;VALUE=DATE:%s\r\nDTEND;VALUE=DATE:%s\r\n", start.Format("20060102"), start.AddDate(0, 0, 1).Format("20060102"))
		} else {
			at := start.Add(time.Duration(e.hour)*time.Hour + time.Duration(e.minute)*time.Minute)
			end := at.Add(time.Duration(max(e.minutes, 15)) * time.Minute)
			fmt.Fprintf(&b, "DTSTART:%s\r\nDTEND:%s\r\n", at.UTC().Format("20060102T150405Z"), end.UTC().Format("20060102T150405Z"))
		}
		if e.location != "" {
			fmt.Fprintf(&b, "LOCATION:%s\r\n", e.location)
		}
		b.WriteString("END:VEVENT\r\n")
	}
	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

// dhlTracking answers the shipment tracking API: the router is out for
// delivery, the books are on their way.
func dhlTracking(req *http.Request, now time.Time) (int, string, string) {
	number := req.URL.Query().Get("trackingNumber")
	code, desc, place, eta := "transit", "The shipment has been loaded onto the delivery vehicle", "Oldenburg", now.Add(3*time.Hour)
	if number == DHLParcels[1].Number {
		code, desc, place, eta = "transit", "The shipment has been processed in the parcel center", "Bremen", now.AddDate(0, 0, 2)
	}
	return jsonOK(mustJSON(map[string]any{"shipments": []any{map[string]any{
		"status": map[string]any{"timestamp": now.Add(-50 * time.Minute).UTC().Format(time.RFC3339), "statusCode": code,
			"status": desc, "description": desc, "location": map[string]any{"address": map[string]string{"addressLocality": place}}},
		"estimatedTimeOfDelivery": eta.UTC().Format(time.RFC3339),
	}}}))
}
