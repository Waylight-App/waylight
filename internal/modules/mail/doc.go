// Package mail is the mail module (#283): unread count and the newest
// unread messages of IMAP inboxes, read-only and headers only.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module; it contains no business logic.
package mail
