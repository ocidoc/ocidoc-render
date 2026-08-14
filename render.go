// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2026 WoozyMasta
// Source: github.com/ocidoc/ocidoc-render

package render

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

// ExternalAssets controls whether HTTP/HTTPS subresources
// (images, audio, video, source, track, poster)
// outside the document itself are left in place or stripped.
// It never affects navigation links (<a href>), only subresource loads.
type ExternalAssets string

const (
	// ExternalAssetsAllow preserves HTTP(S) document subresources.
	ExternalAssetsAllow ExternalAssets = "allow"
	// ExternalAssetsBlock removes HTTP(S) document subresources.
	ExternalAssetsBlock ExternalAssets = "block"
)

// Options controls Markdown and HTML.
type Options struct {
	// DocumentPath is the source document's own bundle-relative path,
	// used to resolve relative local URLs (images, links, media) against its directory.
	// Required: local URL resolution has no other frame of reference.
	DocumentPath string

	// ExternalAssets defaults to ExternalAssetsAllow when empty.
	ExternalAssets ExternalAssets
}

// markdownRenderer enables the GFM extension bundle
// (tables, strikethrough, autolinking and task lists)
// so the Markdown syntax generating them actually reaches documentHTMLPolicy's allowlist,
// which already permits their output elements.
var markdownRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(gmhtml.WithUnsafe()),
)

// Markdown renders Markdown source to a sanitized HTML fragment:
// Goldmark conversion (raw HTML passthrough enabled)
// followed by the same sanitizer and URL/media policy HTML applies directly.
func Markdown(source []byte, opts Options) ([]byte, error) {
	var rendered bytes.Buffer
	if err := markdownRenderer.Convert(source, &rendered); err != nil {
		return nil, fmt.Errorf("render Markdown: %w", err)
	}

	return HTML(rendered.Bytes(), opts)
}

// HTML sanitizes raw HTML source to a safe HTML fragment:
// the explicit document allowlist, then URL classification and media/link normalization.
//
// WithUnsafe (Markdown's own raw-HTML passthrough) does not make its output trusted -
// this is the actual sanitization step,
// and every caller must run it before that output reaches anything.
func HTML(source []byte, opts Options) ([]byte, error) {
	if opts.ExternalAssets == "" {
		opts.ExternalAssets = ExternalAssetsAllow
	}

	fragment := documentHTMLPolicy.SanitizeBytes(source)

	return normalizeDocumentURLs(fragment, opts.DocumentPath, opts.ExternalAssets)
}
