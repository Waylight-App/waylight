// Package electricity is the electricity price module (#288): day-ahead
// market prices per hour for today and tomorrow, with the cheapest hours.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module; it contains no business logic.
package electricity
