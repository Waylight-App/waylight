// Package attention is the attention list module (#599): what needs the
// signed-in account's attention, collected from the widgets on its tabs and
// the connections of the modules behind them.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package attention
