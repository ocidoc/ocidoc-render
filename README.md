# ocidoc-render

`ocidoc-render` renders OCIDoc Markdown
and sanitizes HTML fragments for safe embedding in a user interface.

It accepts GitHub-Flavored Markdown, removes active content,
validates local bundle-relative URLs, normalizes external links
and controls whether external media resources may load.

The package returns an HTML fragment, not a complete HTML document.
Embed its output in a page that supplies the surrounding document structure
and Content Security Policy.

## Install

```sh
go get github.com/ocidoc/ocidoc-render@v0.1.0
```

## Render Markdown

`DocumentPath` is required to resolve local URLs
relative to the OCIDoc document being rendered.

```go
fragment, err := render.Markdown(markdown, render.Options{
    DocumentPath: "docs/README.md",
})
if err != nil {
    return err
}

// Embed fragment only after checking err.
```

Raw Markdown HTML is allowed through Goldmark only so that the sanitizer
can apply one policy to both Markdown-generated and raw HTML.
It does not make the output trusted:
scripts, event handlers, styles, dangerous URL schemes
and disallowed elements are removed.

## Sanitize existing HTML

```go
fragment, err := render.HTML(html, render.Options{
    DocumentPath:   "README.md",
    ExternalAssets: render.ExternalAssetsBlock,
})
if err != nil {
    return err
}
```

`ExternalAssetsBlock` removes HTTP(S) subresource URLs from images,
audio and video while preserving ordinary navigation links.
`ExternalAssetsAllow` is the default and preserves those subresource URLs.

## Scope

This package is intentionally a renderer and sanitizer, not a browser sandbox.
Consumers should still apply an appropriate Content Security Policy,
serve output with a restrictive MIME type
and avoid granting it privileged origin access.

For the OCIDoc format and the full documentation,
see [ocidoc.org](https://ocidoc.org).
