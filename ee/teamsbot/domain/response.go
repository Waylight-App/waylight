// Package domain holds the rules of the Teams command box search (#334):
// what a result looks like and how the answer is built.
//
// Ring: Entities.
package domain

// MaxHits bounds the results shown in the command box.
const MaxHits = 10

// Hit is one search result.
type Hit struct{ Source, Title, URL string }

// Response is the body Teams expects for a composeExtension/query invoke.
type Response struct {
	ComposeExtension ComposeExtension `json:"composeExtension"`
}

// ComposeExtension is the result list or a plain message.
type ComposeExtension struct {
	Type             string       `json:"type"`
	AttachmentLayout string       `json:"attachmentLayout,omitempty"`
	Attachments      []Attachment `json:"attachments,omitempty"`
	Text             string       `json:"text,omitempty"`
}

// Attachment is one result as a thumbnail card.
type Attachment struct {
	ContentType string    `json:"contentType"`
	Content     Thumbnail `json:"content"`
	Preview     *Preview  `json:"preview,omitempty"`
}

// Preview is the card shown in the result list.
type Preview struct {
	ContentType string    `json:"contentType"`
	Content     Thumbnail `json:"content"`
}

// Thumbnail is a thumbnail card.
type Thumbnail struct {
	Title    string   `json:"title"`
	Text     string   `json:"text,omitempty"`
	Buttons  []Action `json:"buttons,omitempty"`
	Tap      *Action  `json:"tap,omitempty"`
	Subtitle string   `json:"subtitle,omitempty"`
}

// Action opens a link.
type Action struct {
	Type  string `json:"type"`
	Title string `json:"title,omitempty"`
	Value string `json:"value"`
}

const thumbnail = "application/vnd.microsoft.card.thumbnail"

// Results builds the answer for hits; none gives a message.
func Results(query string, hits []Hit) Response {
	if len(hits) == 0 {
		return Message("Nothing found for “" + query + "”.")
	}
	out := Response{ComposeExtension: ComposeExtension{Type: "result", AttachmentLayout: "list"}}
	for i, h := range hits {
		if i == MaxHits {
			break
		}
		title := h.Title
		if title == "" {
			title = h.URL
		}
		open := Action{Type: "openUrl", Title: "Open", Value: h.URL}
		out.ComposeExtension.Attachments = append(out.ComposeExtension.Attachments, Attachment{
			ContentType: thumbnail,
			Content:     Thumbnail{Title: title, Subtitle: h.Source, Buttons: []Action{open}, Tap: &open},
			Preview:     &Preview{ContentType: thumbnail, Content: Thumbnail{Title: title, Text: h.Source}},
		})
	}
	return out
}

// Message is a plain text answer.
func Message(text string) Response {
	return Response{ComposeExtension: ComposeExtension{Type: "message", Text: text}}
}

// Texts.
const (
	NotConnected = "Connect your Microsoft 365 account in waylight first, under Settings → Connected accounts."
	Unlicensed   = "The search is not available: the licence key of this waylight does not include chat apps."
	Failed       = "The search did not work this time; try again or use the start page."
)
