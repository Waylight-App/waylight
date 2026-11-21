// Package discovery finds services of the integrations in the home network
// for the widget setup (#590): by Docker container, by usual host name and
// on the gateway, each confirmed by the integration's own probe.
//
// Ring: module wiring. module.go connects the rings below.
package discovery
