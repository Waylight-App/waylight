package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"trisse-home/internal/modules/notifications/domain"
)

// Chat channels (#324, #325) post to the webhook URL kept in Channel.Secret.
// Link is the start page (PUBLIC_URL); empty leaves the link out.

// severityLabel is the short marker in front of a chat message.
func severityLabel(s domain.Severity) string {
	switch s {
	case domain.SeverityCritical:
		return "🔴 Critical"
	case domain.SeverityWarning:
		return "🟠 Warning"
	}
	return "🔵 Info"
}

// slackHeaderLimit is the longest header text Slack accepts.
const slackHeaderLimit = 150

// SlackPublisher posts to a Slack incoming webhook.
type SlackPublisher struct {
	client *http.Client
	link   string
}

// NewSlackPublisher returns a publisher using client.
func NewSlackPublisher(client *http.Client, link string) *SlackPublisher {
	return &SlackPublisher{client: client, link: link}
}

// slackEscape protects &, < and > in mrkdwn, which Slack would otherwise
// read as links or mentions (a service name like "<!channel>" must not ping
// everyone).
func slackEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// SlackMessage builds the Block Kit payload; text is the fallback for
// notifications and clients without blocks.
func SlackMessage(m domain.Message, link string) []byte {
	title := m.Title
	if r := []rune(title); len(r) > slackHeaderLimit {
		title = string(r[:slackHeaderLimit-1]) + "…"
	}
	footer := severityLabel(m.Severity)
	if m.Type != "" {
		footer += " · " + slackEscape(m.Type)
	}
	if link != "" {
		footer += " · <" + link + "|trisse-home>"
	}
	blocks := []map[string]any{
		{"type": "header", "text": map[string]any{"type": "plain_text", "text": title, "emoji": true}},
	}
	if strings.TrimSpace(m.Body) != "" {
		blocks = append(blocks, map[string]any{"type": "section", "text": map[string]any{"type": "mrkdwn", "text": slackEscape(m.Body)}})
	}
	blocks = append(blocks, map[string]any{"type": "context", "elements": []map[string]any{{"type": "mrkdwn", "text": footer}}})
	b, _ := json.Marshal(map[string]any{"text": slackEscape(m.Title + ": " + m.Body), "blocks": blocks}) // plain maps: cannot fail
	return b
}

// Publish implements usecase.Publisher.
func (p *SlackPublisher) Publish(ctx context.Context, c domain.Channel, m domain.Message) error {
	if err := post(ctx, p.client, c.Secret, SlackMessage(m, p.link), nil); err != nil {
		// The error must not contain the URL: it is the credential.
		return fmt.Errorf("slack %w", redact(err, c.Secret))
	}
	return nil
}

// TeamsPublisher posts an Adaptive Card to a Teams Workflows webhook
// ("Post to a channel when a webhook request is received"). The retired
// Office 365 connector webhooks are not supported.
type TeamsPublisher struct {
	client *http.Client
	link   string
}

// NewTeamsPublisher returns a publisher using client.
func NewTeamsPublisher(client *http.Client, link string) *TeamsPublisher {
	return &TeamsPublisher{client: client, link: link}
}

// teamsColor maps severities to Adaptive Card text colours.
func teamsColor(s domain.Severity) string {
	switch s {
	case domain.SeverityCritical:
		return "Attention"
	case domain.SeverityWarning:
		return "Warning"
	}
	return "Default"
}

// adaptiveCardVersion is supported by Teams on desktop, web and mobile.
const adaptiveCardVersion = "1.4"

// TeamsMessage builds the Workflows payload: a message with one Adaptive
// Card attachment, contentUrl null, as Microsoft documents for the trigger.
func TeamsMessage(m domain.Message, link string) []byte {
	sub := severityLabel(m.Severity)
	if m.Type != "" {
		sub += " · " + m.Type
	}
	body := []map[string]any{
		{"type": "TextBlock", "text": m.Title, "weight": "Bolder", "size": "Medium", "wrap": true, "color": teamsColor(m.Severity)},
		{"type": "TextBlock", "text": sub, "isSubtle": true, "spacing": "None", "wrap": true},
	}
	if strings.TrimSpace(m.Body) != "" {
		body = append(body, map[string]any{"type": "TextBlock", "text": m.Body, "wrap": true})
	}
	card := map[string]any{
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"type":    "AdaptiveCard",
		"version": adaptiveCardVersion,
		"body":    body,
	}
	if link != "" {
		card["actions"] = []map[string]any{{"type": "Action.OpenUrl", "title": "Open trisse-home", "url": link}}
	}
	b, _ := json.Marshal(map[string]any{ // plain maps: cannot fail
		"type": "message",
		"attachments": []map[string]any{{
			"contentType": "application/vnd.microsoft.card.adaptive",
			"contentUrl":  nil,
			"content":     card,
		}},
	})
	return b
}

// Publish implements usecase.Publisher.
func (p *TeamsPublisher) Publish(ctx context.Context, c domain.Channel, m domain.Message) error {
	if err := post(ctx, p.client, c.Secret, TeamsMessage(m, p.link), nil); err != nil {
		return fmt.Errorf("teams %w", redact(err, c.Secret))
	}
	return nil
}

// redactedError hides the webhook URL, which net/http includes in
// transport errors, before the error reaches logs or the UI.
type redactedError struct{ msg string }

func (e redactedError) Error() string { return e.msg }

func redact(err error, secret string) error {
	if secret == "" || !strings.Contains(err.Error(), secret) {
		return err
	}
	return redactedError{strings.ReplaceAll(err.Error(), secret, "[webhook URL]")}
}
