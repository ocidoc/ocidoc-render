// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2026 WoozyMasta
// Source: github.com/ocidoc/ocidoc-render

package render

import (
	"net/http"
	"strings"
	"testing"
)

func TestContentSecurityPolicyModes(t *testing.T) {
	t.Parallel()

	allow := ContentSecurityPolicy(ExternalAssetsAllow)
	block := ContentSecurityPolicy(ExternalAssetsBlock)

	wantAllow := "default-src 'none'; script-src 'none'; script-src-attr 'none'; " +
		"style-src 'self'; style-src-attr 'none'; img-src 'self' https: http:; " +
		"media-src 'self' https: http:; font-src 'self'; connect-src 'none'; " +
		"frame-src 'none'; object-src 'none'; worker-src 'none'; base-uri 'none'; " +
		"form-action 'none'; frame-ancestors 'none'"

	wantBlock := "default-src 'none'; script-src 'none'; script-src-attr 'none'; " +
		"style-src 'self'; style-src-attr 'none'; img-src 'self'; media-src 'self'; " +
		"font-src 'self'; connect-src 'none'; frame-src 'none'; object-src 'none'; " +
		"worker-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

	if allow != wantAllow || block != wantBlock {
		t.Fatalf("unexpected CSP values:\nallow: %s\nblock: %s", allow, block)
	}

	if !strings.Contains(allow, "img-src 'self' https: http:") || !strings.Contains(allow, "media-src 'self' https: http:") {
		t.Fatalf("allow CSP does not permit direct HTTP(S) resources: %s", allow)
	}

	if !strings.Contains(block, "img-src 'self'; media-src 'self'") {
		t.Fatalf("block CSP permits external resources: %s", block)
	}

	for _, policy := range []string{allow, block} {
		for _, forbidden := range []string{"'unsafe-inline'", "'unsafe-eval'", "data:", "blob:", " *"} {
			if strings.Contains(policy, forbidden) {
				t.Errorf("CSP contains %q: %s", forbidden, policy)
			}
		}
	}
}

func TestSecurityHeadersSetsExpectedValues(t *testing.T) {
	t.Parallel()

	header := http.Header{}
	SecurityHeaders(header, ExternalAssetsAllow)

	for name, want := range map[string]string{
		"Referrer-Policy":        "no-referrer",
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
	} {
		if got := header.Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}

	if !strings.HasPrefix(header.Get("Content-Security-Policy"), "default-src 'none'") {
		t.Errorf("missing deny-by-default CSP: %s", header.Get("Content-Security-Policy"))
	}
}
