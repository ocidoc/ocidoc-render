// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2026 WoozyMasta
// Source: github.com/ocidoc/ocidoc-render

package render

import (
	"os"
	"path/filepath"
	"testing"
)

var benchmarkFragment []byte

func BenchmarkMarkdown(b *testing.B) {
	source := readBenchmarkFixture(b, "document.md")

	for _, mode := range []ExternalAssets{ExternalAssetsAllow, ExternalAssetsBlock} {
		b.Run(string(mode), func(b *testing.B) {
			opts := Options{
				DocumentPath:   "docs/README.md",
				ExternalAssets: mode,
			}
			b.SetBytes(int64(len(source)))
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				fragment, err := Markdown(source, opts)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkFragment = fragment
			}
		})
	}
}

func BenchmarkHTML(b *testing.B) {
	source := readBenchmarkFixture(b, "document.html")

	for _, mode := range []ExternalAssets{ExternalAssetsAllow, ExternalAssetsBlock} {
		b.Run(string(mode), func(b *testing.B) {
			opts := Options{
				DocumentPath:   "docs/README.md",
				ExternalAssets: mode,
			}
			b.SetBytes(int64(len(source)))
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				fragment, err := HTML(source, opts)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkFragment = fragment
			}
		})
	}
}

func readBenchmarkFixture(t testing.TB, name string) []byte {
	t.Helper()

	path := filepath.Join("testdata", "benchmark", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read benchmark fixture %s: %v", path, err)
	}

	return data
}
