// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2026 WoozyMasta
// Source: github.com/ocidoc/ocidoc-render

package render

import "net/http"

// ContentSecurityPolicy returns the reference deny-by-default CSP
// for a page that embeds this package's sanitized fragment output,
// scoped by mode: under ExternalAssetsBlock, image and media sources are same-origin only
// matching the fragment itself never containing an external subresource URL in that mode.
//
// A caller is free to design its own full-page policy instead -
// this is a ready-to-use default paired with what the sanitizer actually allows through,
// not a requirement.
func ContentSecurityPolicy(mode ExternalAssets) string {
	imgSrc := "'self' https: http:"
	mediaSrc := imgSrc

	if mode == ExternalAssetsBlock {
		imgSrc = "'self'"
		mediaSrc = "'self'"
	}

	return "default-src 'none'; script-src 'none'; script-src-attr 'none'; style-src 'self'; style-src-attr 'none'; " +
		"img-src " + imgSrc + "; media-src " + mediaSrc + "; font-src 'self'; connect-src 'none'; frame-src 'none'; " +
		"object-src 'none'; worker-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"
}

// SecurityHeaders sets the reference response headers
// for a page that embeds this package's sanitized fragment output:
// ContentSecurityPolicy plus Referrer-Policy, X-Content-Type-Options and X-Frame-Options.
//
// As with ContentSecurityPolicy itself, a caller may set these as-is,
// adapt them or build an entirely different policy -
// nothing else in this package depends on them being used.
func SecurityHeaders(header http.Header, mode ExternalAssets) {
	header.Set("Content-Security-Policy", ContentSecurityPolicy(mode))
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
}
