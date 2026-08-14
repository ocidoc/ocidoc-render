// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2026 WoozyMasta
// Source: github.com/ocidoc/ocidoc-render

package render

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/ocidoc/ocidoc-go/spec"
)

// normalizeDocumentURLs re-parses fragment (already sanitized)
// and applies URL/media normalization to every node:
// stripping event handlers, inline style and autoplay, classifying
// and validating every URL attribute, and forcing preload="none"
// on any audio/video element with an external source.
func normalizeDocumentURLs(fragment []byte, documentPath string, mode ExternalAssets) ([]byte, error) {
	contextNode := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}

	nodes, err := html.ParseFragment(bytes.NewReader(fragment), contextNode)
	if err != nil {
		return nil, fmt.Errorf("parse sanitized HTML: %w", err)
	}

	for _, node := range nodes {
		normalizeNode(node, documentPath, mode)
	}

	var output bytes.Buffer
	for _, node := range nodes {
		if err := html.Render(&output, node); err != nil {
			return nil, fmt.Errorf("serialize sanitized HTML: %w", err)
		}
	}

	return output.Bytes(), nil
}

// normalizeNode recurses through node's subtree, normalizing every element's attributes
// and forcing preload="none" on an audio/video element that itself,
// or through a <source> child, references external media.
// It returns whether node's own subtree contains external media,
// so a parent audio/video element can react to a media reference
// nested inside its <source> children.
func normalizeNode(node *html.Node, documentPath string, mode ExternalAssets) bool {
	if node.Type == html.ElementNode {
		normalizeElementAttributes(node, documentPath, mode)
	}

	hasExternalMedia := false
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if normalizeNode(child, documentPath, mode) {
			hasExternalMedia = true
		}
	}

	hasExternalMedia = hasExternalMedia || elementHasExternalMedia(node)
	if node.Type == html.ElementNode && (node.Data == "audio" || node.Data == "video") && hasExternalMedia {
		setAttribute(node, "preload", "none")
	}

	return hasExternalMedia
}

// normalizeElementAttributes strips every event-handler attribute (name begins with "on"),
// inline style and autoplay, then validates every remaining URL attribute, dropping it
// (or, for a resource attribute under ExternalAssetsBlock, dropping it specifically because it is external)
// rather than trusting the sanitizer's element/attribute allowlist alone to have caught a dangerous scheme.
func normalizeElementAttributes(node *html.Node, documentPath string, mode ExternalAssets) {
	for i := 0; i < len(node.Attr); {
		attr := &node.Attr[i]
		if strings.HasPrefix(strings.ToLower(attr.Key), "on") || attr.Key == "style" || attr.Key == "autoplay" {
			node.Attr = append(node.Attr[:i], node.Attr[i+1:]...)
			continue
		}

		kind, isURL := elementURLAttribute(node.Data, attr.Key)
		if !isURL {
			i++
			continue
		}

		urlKind, ok := classifyURL(attr.Val, documentPath, kind == urlResource)
		if !ok || (kind == urlResource && mode == ExternalAssetsBlock && urlKind.external) {
			node.Attr = append(node.Attr[:i], node.Attr[i+1:]...)
			continue
		}

		if node.Data == "a" && urlKind.external {
			setAttribute(node, "rel", "noopener noreferrer")
		}

		i++
	}
}

// urlContext distinguishes a navigational URL from a fetched subresource.
// ExternalAssets applies only to subresources;
// blocking a document's outbound links would change navigation rather than control resource loading.
type urlContext int

const (
	// urlNavigation identifies an anchor href.
	urlNavigation urlContext = iota
	// urlResource identifies an attribute that causes the browser to fetch data.
	urlResource
)

// elementURLAttribute reports whether attribute on element
// is a URL attribute this package understands,
// and whether it is a navigation link (<a href>) or a subresource load (everything else):
// only subresource loads are affected by ExternalAssetsBlock.
func elementURLAttribute(element, attribute string) (urlContext, bool) {
	switch {
	case element == "a" && attribute == "href":
		return urlNavigation, true

	case element == "img" && attribute == "src":
		return urlResource, true

	case (element == "audio" || element == "video" || element == "source" || element == "track") && attribute == "src":
		return urlResource, true

	case element == "video" && attribute == "poster":
		return urlResource, true

	default:
		return urlNavigation, false
	}
}

// classifiedURL describes a URL that passed scheme and bundle-path checks.
type classifiedURL struct {
	// external reports whether the URL needs network access rather than reading from the local artifact.
	external bool
}

// classifyURL validates raw as either a local bundle-relative path or an allowed absolute scheme
// (http/https always; mailto for navigation only, never as a resource source),
// rejecting everything else - including javascript:, data:, blob: and file:,
// which never appear in this function's allowed-scheme switch at all.
//
// A local path is resolved against documentPath's own directory
// when it is not already root-relative, and validated with spec.ValidateBundlePath:
// the same portability rule that decided whether such a path could exist inside an OCIDoc artifact.
func classifyURL(raw, documentPath string, resource bool) (classifiedURL, bool) {
	if raw == "" || raw != strings.TrimSpace(raw) {
		return classifiedURL{}, false
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return classifiedURL{}, false
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		if parsed.Opaque != "" || parsed.Host == "" {
			return classifiedURL{}, false
		}
		return classifiedURL{external: true}, true

	case "mailto":
		if resource || parsed.Opaque == "" {
			return classifiedURL{}, false
		}
		return classifiedURL{external: true}, true

	case "":
		if parsed.Host != "" || parsed.Opaque != "" {
			return classifiedURL{}, false
		}

	default:
		return classifiedURL{}, false
	}

	if parsed.Path == "" {
		return classifiedURL{}, !resource
	}

	localPath := strings.TrimPrefix(parsed.Path, "/")
	if !strings.HasPrefix(parsed.Path, "/") {
		localPath = path.Join(path.Dir(documentPath), parsed.Path)
	}

	if localPath == "" || spec.ValidateBundlePath(localPath) != nil {
		return classifiedURL{}, false
	}

	return classifiedURL{}, true
}

// elementHasExternalMedia reports whether node is an audio/video/source element
// whose own src attribute is an external (http/https) URL.
func elementHasExternalMedia(node *html.Node) bool {
	if node.Type != html.ElementNode || (node.Data != "audio" && node.Data != "video" && node.Data != "source") {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key != "src" {
			continue
		}

		parsed, err := url.Parse(attr.Val)

		return err == nil && (strings.EqualFold(parsed.Scheme, "http") || strings.EqualFold(parsed.Scheme, "https"))
	}

	return false
}

// setAttribute sets key to value on node, replacing any existing value.
func setAttribute(node *html.Node, key, value string) {
	for i := range node.Attr {
		if node.Attr[i].Key == key {
			node.Attr[i].Val = value
			return
		}
	}

	node.Attr = append(node.Attr, html.Attribute{Key: key, Val: value})
}
