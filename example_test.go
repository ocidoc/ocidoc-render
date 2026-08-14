// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2026 WoozyMasta
// Source: github.com/ocidoc/ocidoc-render

package render

import (
	"fmt"
)

func ExampleMarkdown() {
	fragment, err := Markdown([]byte("# Guide\n\n<script>alert(1)</script>"), Options{
		DocumentPath: "docs/README.md",
	})
	if err != nil {
		panic(err)
	}

	fmt.Print(string(fragment))
	// Output:
	// <h1>Guide</h1>
}

func ExampleHTML() {
	fragment, err := HTML([]byte(
		`<p>Read <a href="https://example.com">the guide</a>.</p><img src="https://example.com/logo.svg">`),
		Options{
			DocumentPath:   "README.md",
			ExternalAssets: ExternalAssetsBlock,
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(fragment))
	// Output:
	// <p>Read <a href="https://example.com" rel="noopener noreferrer">the guide</a>.</p><img/>
}
