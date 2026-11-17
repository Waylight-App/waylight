package waste

import (
	"context"
	"fmt"
	"strings"
	"time"

	"waylight/internal/shared/kernel"
)

// SeedDemo implements modules.DemoSeeder: an uploaded collection calendar
// with four bins, so the widget shows the next collections.
func (m *Module) SeedDemo(ctx context.Context) error {
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	if _, configured, err := m.uc.Settings(ctx); err != nil || configured {
		return err
	}
	_, err := m.uc.UploadCalendar(ctx, "collection-calendar.ics", demoCalendar(time.Now()))
	return err
}

// demoCalendar is an ICS file with repeating all-day collections that start
// in the next days.
func demoCalendar(now time.Time) []byte {
	bins := []struct {
		summary string
		inDays  int
		weeks   int
	}{
		{"Organic waste", 1, 1},
		{"Residual waste", 3, 2},
		{"Recycling (yellow bag)", 6, 2},
		{"Paper", 10, 4},
	}
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Waylight//Demo//EN\r\nX-WR-CALNAME:Collection calendar\r\n")
	for i, bin := range bins {
		start := now.AddDate(0, 0, bin.inDays)
		fmt.Fprintf(&b, "BEGIN:VEVENT\r\nUID:demo-waste-%d@waylight.example\r\nDTSTAMP:%s\r\n", i, now.UTC().Format("20060102T150405Z"))
		fmt.Fprintf(&b, "DTSTART;VALUE=DATE:%s\r\nDTEND;VALUE=DATE:%s\r\n", start.Format("20060102"), start.AddDate(0, 0, 1).Format("20060102"))
		fmt.Fprintf(&b, "RRULE:FREQ=WEEKLY;INTERVAL=%d;COUNT=26\r\nSUMMARY:%s\r\nEND:VEVENT\r\n", bin.weeks, bin.summary)
	}
	b.WriteString("END:VCALENDAR\r\n")
	return []byte(b.String())
}
