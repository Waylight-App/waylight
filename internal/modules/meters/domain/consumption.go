package domain

import "time"

// Amount is a consumption; Known is false when the readings do not cover the
// period.
type Amount struct {
	Value float64
	Known bool
}

// YearConsumption is the consumption of one calendar year.
type YearConsumption struct {
	Year   int
	Months [12]Amount
	// Total covers the part of the year between the first and the last
	// reading; Complete says it is the whole year.
	Total    Amount
	Complete bool
}

// valueOn returns the counter value on day, interpolated linearly between
// the two readings around it. readings must be sorted by day; outside the
// first and the last reading the value is unknown.
func valueOn(readings []Reading, day time.Time) (float64, bool) {
	for i, r := range readings {
		if r.Day.Equal(day) {
			return r.Value, true
		}
		if r.Day.After(day) {
			if i == 0 {
				return 0, false
			}
			prev := readings[i-1]
			span := r.Day.Sub(prev.Day).Hours()
			done := day.Sub(prev.Day).Hours()
			return prev.Value + (r.Value-prev.Value)*done/span, true
		}
	}
	return 0, false
}

// between returns the consumption from day from up to day to.
func between(readings []Reading, from, to time.Time) Amount {
	a, okA := valueOn(readings, from)
	b, okB := valueOn(readings, to)
	if !okA || !okB {
		return Amount{}
	}
	return Amount{Value: b - a, Known: true}
}

// ConsumptionOf computes the consumption per month and for the year from
// readings sorted by day. A month is known when the readings reach from
// its first day to the first day of the next month.
func ConsumptionOf(readings []Reading, year int) YearConsumption {
	out := YearConsumption{Year: year}
	for m := range out.Months {
		from := time.Date(year, time.Month(m+1), 1, 0, 0, 0, 0, time.UTC)
		out.Months[m] = between(readings, from, from.AddDate(0, 1, 0))
	}
	if len(readings) == 0 {
		return out
	}
	start, end := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	from, to := start, end
	if first := readings[0].Day; first.After(from) {
		from = first
	}
	if last := readings[len(readings)-1].Day; last.Before(to) {
		to = last
	}
	out.Complete = from.Equal(start) && to.Equal(end)
	if from.Before(to) {
		out.Total = between(readings, from, to)
	}
	return out
}

// Years lists the calendar years the readings touch, the latest first.
func Years(readings []Reading) []int {
	if len(readings) == 0 {
		return nil
	}
	first, last := readings[0].Day.Year(), readings[len(readings)-1].Day.Year()
	out := make([]int, 0, last-first+1)
	for y := last; y >= first; y-- {
		out = append(out, y)
	}
	return out
}
