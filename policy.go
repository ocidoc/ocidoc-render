// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2026 WoozyMasta
// Source: github.com/ocidoc/ocidoc-render

package render

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

// documentHTMLPolicy is the complete v1beta passive-document HTML allowlist,
// built once and reused: an explicit element/attribute list rather than bluemonday.UGCPolicy()
// whose exact contents are not the OCIDoc contract.
// URL attributes are checked again by normalizeDocumentURLs -
// this policy alone would still accept a dangerous scheme on an allowed attribute.
var documentHTMLPolicy = newDocumentHTMLPolicy()

// newDocumentHTMLPolicy constructs the closed allowlist for passive document content.
// URL safety is enforced separately because an allowed attribute alone
// does not make its value safe to load or navigate to.
func newDocumentHTMLPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements(
		"article", "section", "div", "span", "p", "hr", "br",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"strong", "b", "em", "i", "del", "s", "sub", "sup", "small", "mark",
		"kbd", "samp", "var", "q", "cite", "abbr", "pre", "code",
		"ul", "ol", "li", "dl", "dt", "dd", "blockquote", "details", "summary",
		"table", "caption", "thead", "tbody", "tfoot", "tr", "th", "td", "colgroup", "col",
		"figure", "figcaption", "a", "img", "audio", "video", "source", "track",
	)
	p.AllowAttrs("class", "id", "title", "dir", "lang").Globally()
	p.AllowAttrs("width", "height", "align").OnElements("img", "audio", "video", "table", "th", "td", "col")
	p.AllowAttrs("colspan", "rowspan").OnElements("th", "td")
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src", "alt").OnElements("img")
	p.AllowAttrs("src", "controls", "loop", "muted", "playsinline").OnElements("audio", "video")
	p.AllowAttrs("poster").OnElements("video")
	p.AllowAttrs("src", "type", "media").OnElements("source")
	p.AllowAttrs("src", "kind", "srclang", "label", "default").OnElements("track")
	p.AllowAttrs("preload").Matching(regexp.MustCompile(`^(?:none|metadata)$`)).OnElements("audio", "video")
	p.AllowRelativeURLs(true)
	p.AllowURLSchemes("http", "https", "mailto")

	return p
}
