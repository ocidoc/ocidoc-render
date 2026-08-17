// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2026 WoozyMasta
// Source: github.com/ocidoc/ocidoc-render

package render

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var alertMarkerPattern = regexp.MustCompile(`(?i)^\[!(note|tip|important|warning|caution)\](?:[ \t]+(.*))?$`)

// alertASTTransformer turns GitHub and GitLab-compatible alert blockquotes
// into regular blockquotes with safe CSS classes and a visible title paragraph.
type alertASTTransformer struct{}

// Transform applies alert metadata to blockquotes in the Markdown AST.
func (alertASTTransformer) Transform(document *ast.Document, reader text.Reader, _ parser.Context) {
	source := reader.Source()
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || node.Kind() != ast.KindBlockquote {
			return ast.WalkContinue, nil
		}

		blockquote := node.(*ast.Blockquote)
		paragraph, ok := blockquote.FirstChild().(*ast.Paragraph)
		if !ok {
			return ast.WalkContinue, nil
		}
		if _, ok := paragraph.FirstChild().(*ast.Text); !ok {
			return ast.WalkContinue, nil
		}

		markerText := string(paragraph.Lines().Value(source))
		line := markerText
		if lineEnd := strings.IndexByte(line, '\n'); lineEnd >= 0 {
			line = line[:lineEnd]
		}
		match := alertMarkerPattern.FindStringSubmatch(strings.TrimSpace(line))
		if match == nil {
			return ast.WalkContinue, nil
		}

		kind := strings.ToLower(match[1])
		title := match[2]
		if title == "" {
			title = strings.ToUpper(kind[:1]) + kind[1:]
		}

		blockquote.SetAttributeString("class", "markdown-alert markdown-alert-"+kind)
		removeAlertMarker(paragraph, len(line), source)

		titleParagraph := ast.NewParagraph()
		titleParagraph.SetAttributeString("class", "markdown-alert-title")
		titleParagraph.AppendChild(titleParagraph, ast.NewString([]byte(title)))
		blockquote.InsertBefore(blockquote, paragraph, titleParagraph)

		return ast.WalkContinue, nil
	})
}

// removeAlertMarker removes the alert marker line while preserving the remaining inline AST.
func removeAlertMarker(paragraph *ast.Paragraph, length int, source []byte) {
	remaining := length
	for child := paragraph.FirstChild(); child != nil && remaining > 0; {
		next := child.NextSibling()
		if textNode, ok := child.(*ast.Text); ok {
			valueLength := len(textNode.Value(source))
			if valueLength <= remaining {
				paragraph.RemoveChild(paragraph, child)
				remaining -= valueLength
			} else {
				textNode.Segment = textNode.Segment.WithStart(textNode.Segment.Start + remaining)
				remaining = 0
			}
		} else {
			paragraph.RemoveChild(paragraph, child)
		}
		child = next
	}
}
