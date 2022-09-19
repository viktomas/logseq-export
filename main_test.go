package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var rawContent = `public:: true
title:: Blog article: hello & world

- First paragraph
- Second paragraph
	- bullet point
`

var textPart = `- First paragraph
- Second paragraph
	- bullet point
`

func TestParseTextAndAttributes(t *testing.T) {
	text, attributes := parseTextAndAttributes(rawContent)
	assert.Equal(t, textPart, text)
	assert.Equal(t, map[string]string{
		"public": "true",
		"title":  "Blog article: hello & world",
	}, attributes)
}

func TestParsePage(t *testing.T) {
	parsedPage := parsePage("Blog article%3A hello %26 world", rawContent)
	assert.Equal(t, page{
		filename: "Blog article%3A hello %26 world",
		attributes: map[string]string{
			"public": "true",
			"title":  "Blog article: hello & world",
		},
		text: textPart,
	}, parsedPage)
}
