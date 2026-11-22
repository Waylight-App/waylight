// Package setupsignin signs in to a service from the widget setup assistant
// instead of asking for a pasted key (#654): Nextcloud Login Flow v2 and
// Home Assistant's OAuth. The credential stays on the server; the module
// that is set up redeems it once. It also carries a key copied on the phone
// to the assistant (Paste from phone): a short-lived, single-use code.
//
// Ring: module wiring. module.go connects the rings below.
package setupsignin
