// Package domain holds the rules of the Slack command (#333): what a
// search result looks like in Slack and how the reply is written.
//
// Ring: Entities.
package domain

import (
	"strings"
	"unicode/utf8"
)

// MaxHits bounds the results in one reply; Slack is not the place for a
// long list, the start page is.
const MaxHits = 8

// Hit is one search result.
type Hit struct {
	// Source is shown after the link: "Service", "Bookmark", or the name of
	// the wiki it came from.
	Source string
	Title  string
	URL    string
}

// escape makes text safe in Slack's mrkdwn, where &, < and > start markup.
func escape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// linkText keeps a link label short and free of the characters that end it.
func linkText(s string) string {
	s = strings.NewReplacer("|", "¦", "\n", " ").Replace(strings.TrimSpace(s))
	if utf8.RuneCountInString(s) > 80 {
		s = string([]rune(s)[:79]) + "…"
	}
	return escape(s)
}

// Results writes the reply for a search.
func Results(query string, hits []Hit, more bool) string {
	q := escape(query)
	if len(hits) == 0 {
		return "Nothing found for “" + q + "”."
	}
	var b strings.Builder
	b.WriteString("Found for “" + q + "”:")
	for i, h := range hits {
		if i == MaxHits {
			break
		}
		title := h.Title
		if title == "" {
			title = h.URL
		}
		b.WriteString("\n• <" + strings.ReplaceAll(h.URL, ">", "%3E") + "|" + linkText(title) + ">")
		if h.Source != "" {
			b.WriteString(" · " + escape(h.Source))
		}
	}
	if more || len(hits) > MaxHits {
		b.WriteString("\nMore on the start page.")
	}
	return b.String()
}

// Help is the reply to the command without words.
func Help(command string) string {
	c := escape(command)
	return "Search the start page: " + c + " vpn, " + c + " holiday rules. Only you see the answer."
}

// NotConnected is the reply to someone whose Slack account is not linked to
// an account of the instance.
func NotConnected(connectURL string) string {
	msg := "Connect your Slack account in trisse-home first, under Settings → Connected accounts."
	if connectURL != "" {
		msg += " <" + connectURL + "|Open the settings>"
	}
	return msg
}

// Unlicensed is the reply when the licence does not cover chat apps.
const Unlicensed = "The Slack command is not available: the licence key of this trisse-home does not include chat apps."
