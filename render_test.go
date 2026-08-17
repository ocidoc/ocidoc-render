// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2026 WoozyMasta
// Source: github.com/ocidoc/ocidoc-render

package render

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestMarkdownSanitizesActiveContent(t *testing.T) {
	t.Parallel()

	const tagStart = "\x3c"

	source := []byte(strings.Join([]string{
		"# Safe",
		tagStart + "script>alert(1)" + tagStart + "/script>",
		tagStart + `img src="assets/image.svg" onerror="alert(1)" alt="safe">`,
		tagStart + `a href="javascript:alert(1)">bad` + tagStart + "/a>",
		tagStart + `div onclick="alert(1)" style="color:red">content` + tagStart + "/div>",
		tagStart + `iframe src="https://example.com">` + tagStart + "/iframe>",
		tagStart + `object data="x">` + tagStart + "/object>" + tagStart + `embed src="x">`,
		tagStart + "style>body { display: none }" + tagStart + "/style>",
		tagStart + `svg onload="alert(1)">` + tagStart + "circle />" + tagStart + "/svg>",
		"",
	}, "\n"))

	fragment, err := Markdown(source, Options{DocumentPath: "docs/README.md"})
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}

	got := string(fragment)
	for _, forbidden := range []string{
		"<script", "onerror", "onclick", "javascript:",
		"<iframe", "<object", "<embed", "<style", "style=", "<svg",
	} {
		if strings.Contains(strings.ToLower(got), forbidden) {
			t.Errorf("rendered fragment contains forbidden %q: %s", forbidden, got)
		}
	}

	for _, allowed := range []string{"<h1>Safe</h1>", `<img src="assets/image.svg" alt="safe"/>`, "<div>content</div>"} {
		if !strings.Contains(got, allowed) {
			t.Errorf("rendered fragment does not contain allowed %q: %s", allowed, got)
		}
	}
}

func TestMarkdownNormalizesMediaAndExternalLinks(t *testing.T) {
	t.Parallel()

	source := []byte(`<a href="https://example.com/page">external</a>
<audio controls autoplay preload="metadata" src="http://media.example/demo.ogg" onplay="x"></audio>
<video controls autoplay poster="https://media.example/poster.png">
  <source src="https://media.example/demo.webm" type="video/webm">
  <source src="assets/demo.mp4" type="video/mp4">
  <track kind="subtitles" src="assets/demo.vtt" srclang="en" label="English" default>
</video>`)

	fragment, err := Markdown(source, Options{DocumentPath: "README.md"})
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}

	got := string(fragment)
	for _, want := range []string{
		`href="https://example.com/page" rel="noopener noreferrer"`,
		`<audio controls="" preload="none" src="http://media.example/demo.ogg"></audio>`,
		`poster="https://media.example/poster.png"`,
		`src="https://media.example/demo.webm"`,
		`src="assets/demo.mp4"`,
		`<track kind="subtitles" src="assets/demo.vtt" srclang="en" label="English" default=""/>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered fragment does not contain %q: %s", want, got)
		}
	}

	for _, forbidden := range []string{"autoplay", "onplay"} {
		if strings.Contains(got, forbidden) {
			t.Errorf("rendered fragment contains %q: %s", forbidden, got)
		}
	}
}

func TestMarkdownRendersFootnotes(t *testing.T) {
	t.Parallel()

	fragment, err := Markdown([]byte("Text with a footnote[^1].\n\n[^1]: Footnote text.\n"), Options{
		DocumentPath: "README.md",
	})
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}

	got := string(fragment)
	for _, want := range []string{"Footnote text.", "footnote-ref", "footnotes"} {
		if !strings.Contains(got, want) {
			t.Errorf("footnote output does not contain %q: %s", want, got)
		}
	}
}

func TestMarkdownRendersAlerts(t *testing.T) {
	t.Parallel()

	fragment, err := Markdown([]byte(
		"> [!warning] Data deletion\n> This **cannot** be undone.\n\n> [!TIP]\n> Keep a backup.\n",
	), Options{DocumentPath: "README.md"})
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}

	got := string(fragment)
	for _, want := range []string{
		`<blockquote class="markdown-alert markdown-alert-warning">`,
		`<p class="markdown-alert-title">Data deletion</p>`,
		`<strong>cannot</strong>`,
		`<blockquote class="markdown-alert markdown-alert-tip">`,
		`<p class="markdown-alert-title">Tip</p>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("alert output does not contain %q: %s", want, got)
		}
	}
}

func TestMarkdownBlocksOnlyExternalSubresources(t *testing.T) {
	t.Parallel()

	source := []byte(`<a href="https://example.com/page">external link</a>
<img src="https://example.com/image.png"><img src="assets/local.png">
<video src="https://example.com/video.mp4" poster="poster.png"></video>`)

	fragment, err := Markdown(
		source,
		Options{DocumentPath: "docs/README.md", ExternalAssets: ExternalAssetsBlock},
	)
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}

	got := string(fragment)
	if !strings.Contains(got, `href="https://example.com/page"`) ||
		!strings.Contains(got, `src="assets/local.png"`) ||
		!strings.Contains(got, `poster="poster.png"`) {
		t.Fatalf("block mode removed allowed navigation/local resources: %s", got)
	}

	if strings.Contains(got, `src="https://`) {
		t.Fatalf("block mode retained external subresource: %s", got)
	}
}

// TestHTML exercises HTML directly (not through Markdown),
// asserting an exact rendered fragment for every case rather than checking
// for the presence or absence of a substring:
// a case that should be removed must produce exactly its safe remainder,
// and a case that should be preserved must round-trip byte-for-byte,
// not merely "contain" its allowed tag somewhere in a larger, unchecked string.
func TestHTML(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"script-stripped", "event-handler-and-inline-style-stripped",
		"javascript-href-stripped-text-kept", "iframe-removed-entirely",
		"malformed-markup-escaped-not-executed", "heading-with-id-and-class",
		"inline-semantic-elements", "unordered-list", "description-list",
		"table-with-colspan", "local-link-kept-plain", "external-link-gets-rel",
		"image-with-dimensions", "audio-with-preload-kept-as-is",
		"video-with-source-and-track", "details-and-summary",
		"figure-and-figcaption", "blockquote",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			input := readHTMLFixture(t, name, "input.html")
			want := readHTMLFixture(t, name, "want.html")
			got, err := HTML(input, Options{DocumentPath: "docs/README.md"})
			if err != nil {
				t.Fatalf("HTML: %v", err)
			}
			if string(got) != string(want) {
				t.Errorf("HTML(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

func readHTMLFixture(t testing.TB, name, suffix string) []byte {
	t.Helper()

	path := filepath.Join("testdata", "html", name+"."+suffix)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}

	return bytes.TrimSuffix(data, []byte("\n"))
}

func TestClassifyURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      string
		resource bool
		ok       bool
		external bool
	}{
		{name: "relative", raw: "../assets/a.png", resource: true, ok: true},
		{name: "root relative", raw: "/assets/a.png", resource: true, ok: true},
		{name: "fragment", raw: "#section", ok: true},
		{name: "http", raw: "http://example.com/a", resource: true, ok: true, external: true},
		{name: "https", raw: "https://example.com/a", resource: true, ok: true, external: true},
		{name: "mailto link", raw: "mailto:user@example.com", ok: true, external: true},
		{name: "mailto resource", raw: "mailto:user@example.com", resource: true},
		{name: "javascript", raw: "javascript:alert(1)"},
		{name: "data", raw: "data:image/png;base64,eA==", resource: true},
		{name: "file", raw: "file:///etc/passwd"},
		{name: "blob", raw: "blob:https://example.com/id", resource: true},
		{name: "protocol relative", raw: "//example.com/a", resource: true},
		{name: "escape", raw: "../../secret", resource: true},
		{name: "backslash", raw: `..\secret`, resource: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := classifyURL(test.raw, "docs/README.md", test.resource)
			if ok != test.ok || got.external != test.external {
				t.Fatalf("classifyURL(%q) = (%+v, %v), want external=%v ok=%v", test.raw, got, ok, test.external, test.ok)
			}
		})
	}
}

func FuzzClassifyURL(f *testing.F) {
	for _, seed := range []string{
		"assets/a.png", "../a", "javascript:alert(1)", "https://example.com/a", "data:text/html,x", "\x00",
	} {
		f.Add(seed, true)
		f.Add(seed, false)
	}

	f.Fuzz(func(t *testing.T, raw string, resource bool) {
		got, ok := classifyURL(raw, "docs/README.md", resource)
		assertClassifiedURLInvariant(t, raw, resource, got, ok)
	})
}

func FuzzHTML(f *testing.F) {
	for _, name := range []string{
		"safe-paragraph",
		"script",
		"image-onerror",
		"javascript-link",
		"svg-onload",
		"malformed",
	} {
		f.Add(string(readHTMLFixture(f, "fuzz/"+name, "html")))
	}

	f.Fuzz(func(t *testing.T, source string) {
		for _, mode := range []ExternalAssets{ExternalAssetsAllow, ExternalAssetsBlock} {
			assertSafeFragment(t, func() ([]byte, error) {
				return HTML([]byte(source), Options{
					DocumentPath:   "README.html",
					ExternalAssets: mode,
				})
			}, "README.html", mode)
		}
	})
}

func FuzzMarkdown(f *testing.F) {
	f.Add(string(readBenchmarkFixture(f, "document.md")))
	f.Add("# Safe document\n\nLocal [link](assets/guide.md).")

	f.Fuzz(func(t *testing.T, source string) {
		for _, mode := range []ExternalAssets{ExternalAssetsAllow, ExternalAssetsBlock} {
			assertSafeFragment(t, func() ([]byte, error) {
				return Markdown([]byte(source), Options{
					DocumentPath:   "docs/README.md",
					ExternalAssets: mode,
				})
			}, "docs/README.md", mode)
		}
	})
}

func assertSafeFragment(t *testing.T, render func() ([]byte, error), documentPath string, mode ExternalAssets) {
	t.Helper()

	fragment, err := render()
	if err != nil {
		t.Fatalf("render fragment: %v", err)
	}
	contextNode := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(bytes.NewReader(fragment), contextNode)
	if err != nil {
		t.Fatalf("parse rendered fragment: %v", err)
	}
	for _, node := range nodes {
		assertNoActiveHTML(t, node, documentPath, mode)
	}
}

func assertNoActiveHTML(t *testing.T, node *html.Node, documentPath string, mode ExternalAssets) {
	t.Helper()

	if node.Type == html.ElementNode {
		switch node.Data {
		case "script", "iframe", "frame", "frameset", "object", "embed", "applet", "base", "link", "form", "style", "svg", "math":
			t.Fatalf("active element %q survived sanitization", node.Data)
		}

		for _, attr := range node.Attr {
			if strings.HasPrefix(strings.ToLower(attr.Key), "on") || attr.Key == "style" || attr.Key == "autoplay" {
				t.Fatalf("active attribute %q survived sanitization", attr.Key)
			}

			context, isURL := elementURLAttribute(node.Data, attr.Key)
			if isURL {
				assertRenderedURLInvariant(t, node, attr, documentPath, mode, context)
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		assertNoActiveHTML(t, child, documentPath, mode)
	}
}

func assertClassifiedURLInvariant(t *testing.T, raw string, resource bool, got classifiedURL, ok bool) {
	t.Helper()
	if !ok {
		return
	}
	if raw == "" || raw != strings.TrimSpace(raw) {
		t.Fatalf("accepted empty or whitespace-padded URL %q", raw)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("accepted unparsable URL %q: %v", raw, err)
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		if parsed.Opaque != "" || parsed.Host == "" || !got.external {
			t.Fatalf("accepted invalid external URL %q as %+v", raw, got)
		}

	case "mailto":
		if resource || parsed.Opaque == "" || !got.external {
			t.Fatalf("accepted invalid mailto URL %q as %+v", raw, got)
		}

	case "":
		if parsed.Host != "" || parsed.Opaque != "" || got.external || (resource && parsed.Path == "") {
			t.Fatalf("accepted invalid local URL %q as %+v", raw, got)
		}

	default:
		t.Fatalf("accepted unsupported URL scheme in %q", raw)
	}
}

func assertRenderedURLInvariant(
	t *testing.T,
	node *html.Node,
	attr html.Attribute,
	documentPath string,
	mode ExternalAssets,
	context urlContext,
) {
	t.Helper()

	got, ok := classifyURL(attr.Val, documentPath, context == urlResource)
	if !ok {
		t.Fatalf("unsafe URL attribute survived: <%s %s=%q>", node.Data, attr.Key, attr.Val)
	}
	assertClassifiedURLInvariant(t, attr.Val, context == urlResource, got, ok)
	if context == urlResource && mode == ExternalAssetsBlock && got.external {
		t.Fatalf("external resource survived block mode: <%s %s=%q>", node.Data, attr.Key, attr.Val)
	}
	if node.Data == "a" && got.external && !hasNoopenerNoreferrer(node) {
		t.Fatalf("external link lacks rel=noopener noreferrer: %q", attr.Val)
	}
}

func hasNoopenerNoreferrer(node *html.Node) bool {
	for _, attr := range node.Attr {
		if attr.Key != "rel" {
			continue
		}
		words := strings.Fields(attr.Val)
		return slices.Contains(words, "noopener") && slices.Contains(words, "noreferrer")
	}

	return false
}
