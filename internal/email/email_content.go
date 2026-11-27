package email

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// Content is what an email says. One layout renders every kind as HTML and
// as plain text, so all emails look alike and stay readable everywhere.
type Content struct {
	Subject    string   `json:"subject"`
	Preheader  string   `json:"preheader,omitempty"` // preview line in inbox lists
	Heading    string   `json:"heading"`
	Paragraphs []string `json:"paragraphs,omitempty"`
	Code       string   `json:"code,omitempty"` // shown large, e.g. a one-time password
	CodeNote   string   `json:"code_note,omitempty"`
	Details    []Detail `json:"details,omitempty"`
	Quote      string   `json:"quote,omitempty"` // e.g. a company's reason
	Button     *Button  `json:"button,omitempty"`
	After      []string `json:"after,omitempty"` // paragraphs after the button
}

type Detail struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Button struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Draft is an email to queue: who gets it, a kind for logs, and the content.
type Draft struct {
	Kind    string
	To      string
	Content Content
}

// Text renders the plain-text version.
func (c Content) Text() string {
	var b strings.Builder
	b.WriteString(c.Heading + "\n\n")
	for _, p := range c.Paragraphs {
		b.WriteString(p + "\n\n")
	}
	if c.Code != "" {
		b.WriteString("    " + c.Code + "\n\n")
		if c.CodeNote != "" {
			b.WriteString(c.CodeNote + "\n\n")
		}
	}
	if len(c.Details) > 0 {
		w := 0
		for _, d := range c.Details {
			if len(d.Label) > w {
				w = len(d.Label)
			}
		}
		for _, d := range c.Details {
			b.WriteString(fmt.Sprintf("%-*s  %s\n", w+1, d.Label+":", d.Value))
		}
		b.WriteString("\n")
	}
	if c.Quote != "" {
		b.WriteString("\"" + c.Quote + "\"\n\n")
	}
	if c.Button != nil {
		b.WriteString(c.Button.Label + ": " + c.Button.URL + "\n\n")
	}
	for _, p := range c.After {
		b.WriteString(p + "\n\n")
	}
	b.WriteString("The SafariQuest team\n")
	return b.String()
}

// HTML renders the HTML version. Styles are inline and the layout is a
// table, which is what email clients reliably support.
func (c Content) HTML() (string, error) {
	var b bytes.Buffer
	err := layout.Execute(&b, c)
	return b.String(), err
}

var layout = template.Must(template.New("email").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Subject}}</title></head>
<body style="margin:0;padding:0;background:#f4efe8;font-family:-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#1c1b18;">
{{if .Preheader}}<div style="display:none;max-height:0;overflow:hidden;opacity:0;">{{.Preheader}}</div>{{end}}
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#f4efe8;padding:24px 12px;">
<tr><td align="center">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:560px;background:#ffffff;border-radius:16px;overflow:hidden;">
<tr><td style="background:#713700;padding:22px 28px;color:#ffffff;font-size:20px;font-weight:700;letter-spacing:-0.01em;">SafariQuest</td></tr>
<tr><td style="padding:28px 28px 8px;">
<h1 style="margin:0 0 16px;font-size:22px;line-height:1.3;font-weight:600;color:#1c1b18;">{{.Heading}}</h1>
{{range .Paragraphs}}<p style="margin:0 0 14px;font-size:15px;line-height:1.55;color:#3a3732;">{{.}}</p>{{end}}
{{if .Code}}
<div style="margin:18px 0 8px;padding:18px;border-radius:12px;background:#f7f3ed;text-align:center;font-family:SFMono-Regular,Menlo,Consolas,monospace;font-size:34px;font-weight:700;letter-spacing:10px;color:#1c1b18;">{{.Code}}</div>
{{if .CodeNote}}<p style="margin:0 0 14px;font-size:13px;line-height:1.5;color:#6e6860;text-align:center;">{{.CodeNote}}</p>{{end}}
{{end}}
{{if .Details}}
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="margin:14px 0 16px;border-top:1px solid #ece5db;">
{{range .Details}}<tr>
<td style="padding:9px 0;border-bottom:1px solid #ece5db;font-size:13px;color:#6e6860;width:38%;vertical-align:top;">{{.Label}}</td>
<td style="padding:9px 0;border-bottom:1px solid #ece5db;font-size:14px;color:#1c1b18;vertical-align:top;">{{.Value}}</td>
</tr>{{end}}
</table>
{{end}}
{{if .Quote}}<p style="margin:0 0 16px;padding:12px 16px;border-left:3px solid #713700;background:#fbf7f2;font-size:14px;line-height:1.5;font-style:italic;color:#3a3732;">“{{.Quote}}”</p>{{end}}
{{if .Button}}
<table role="presentation" cellpadding="0" cellspacing="0" style="margin:8px 0 20px;"><tr><td style="border-radius:999px;background:#1c1b18;">
<a href="{{.Button.URL}}" style="display:inline-block;padding:12px 22px;font-size:14px;font-weight:600;color:#ffffff;text-decoration:none;">{{.Button.Label}}</a>
</td></tr></table>
{{end}}
{{range .After}}<p style="margin:0 0 14px;font-size:14px;line-height:1.55;color:#3a3732;">{{.}}</p>{{end}}
</td></tr>
<tr><td style="padding:16px 28px 26px;font-size:12px;line-height:1.5;color:#8a837a;border-top:1px solid #ece5db;">
The SafariQuest team · Ethical safaris and cultural immersion in Tanzania
</td></tr>
</table>
</td></tr></table>
</body></html>`))
