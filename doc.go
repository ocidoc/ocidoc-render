// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2026 WoozyMasta
// Source: github.com/ocidoc/ocidoc-render

/*
Package render converts Markdown or raw HTML into a sanitized HTML fragment
safe to embed in a trusted page template.

It owns the reusable, document-format-agnostic part of that security boundary:
Markdown-to-HTML conversion, the explicit allowlist sanitizer,
dangerous URL scheme rejection, and native audio/video/source/track handling.
Markdown and HTML both return a fragment, not a complete HTML document;
routing, local asset serving and the trusted outer page template
stay the hosting application's own job,
since those are inherently specific to how each application is built and served.
ContentSecurityPolicy and SecurityHeaders provide a ready-to-use reference policy
for the one full-page concern that pairs directly with what the sanitizer allows through,
so a consumer is not forced to invent an equivalent CSP from scratch,
but using them is optional.

This package has no dependency on OCIDoc registry, store or CLI packages;
it depends on github.com/ocidoc/ocidoc-go/spec only for spec.ValidateBundlePath,
the same portable bundle-path rule OCIDoc artifacts already use,
so a local URL is judged "valid" by the exact rule that decided whether the file
it points to could exist in an OCIDoc artifact in the first place.
*/
package render
