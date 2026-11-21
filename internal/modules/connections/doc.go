// Package connections is the Connections module (#610): named connections
// to a service, saved once with an encrypted credential and chosen by the
// widgets that use the service.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package connections
