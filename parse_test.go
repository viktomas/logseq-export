package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestParseAttributes(t *testing.T) {
	attributes := parseAttributes(rawContent)
	assert.Equal(t, map[string]string{
		"public": "true",
		"title":  "Blog article: hello & world",
	}, attributes)
}

func TestStripAttributes(t *testing.T) {
	textWithoutAttributes := stripAttributes(rawContent)
	require.Equal(t, textPart, textWithoutAttributes)
}

func TestParsePage(t *testing.T) {
	parsedPage := parsePageOld("Blog article%3A hello %26 world", rawContent)
	assert.Equal(t, oldPage{
		filename: "Blog article%3A hello %26 world",
		attributes: map[string]string{
			"public": "true",
			"title":  "Blog article: hello & world",
		},
		text: textPart,
	}, parsedPage)
}

func TestParseAssets(t *testing.T) {
	t.Run("extracts relative images", func(t *testing.T) {
		content := "- ![hello world](../assets/image.png)"
		result := parseAssets(content)
		require.Equal(t, []string{"../assets/image.png"}, result)
	})

	t.Run("ignores absolute images", func(t *testing.T) {
		content := "- ![hello world](http://example.com/assets/image.png)"
		result := parseAssets(content)
		require.Equal(t, 0, len(result))
	})

	// TODO if first content line contains only image, move it to an image attribute (based on some config)
	// t.Run("extracts relative images from image attribute", func(t *testing.T) {
	// 	testPage := oldPage{
	// 		attributes: map[string]string{
	// 			"image": "../assets/image.png",
	// 		},
	// 		filename: "a.md",
	// 		text:     "",
	// 	}
	// 	result := transformPage(testPage, "/images")
	// 	require.Equal(t, []string{"../assets/image.png"}, result.assets)
	// 	require.Equal(t, "/images/image.png", result.attributes["image"])
	// })

	// TODO
	// t.Run("ignores absolute images in image attribute", func(t *testing.T) {
	// 	testPage := oldPage{
	// 		attributes: map[string]string{
	// 			"image": "http://example.com/assets/image.png",
	// 		},
	// 		filename: "a.md",
	// 		text:     "",
	// 	}
	// 	result := transformPage(testPage, "/images")
	// 	require.Equal(t, 0, len(result.assets))
	// 	require.Equal(t, "http://example.com/assets/image.png", result.attributes["image"])
	// })

}

func TestParseText(t *testing.T) {
	t.Run("removes dashes with no text after them", func(t *testing.T) {
		result := parseContent("-\n\t- \n\t\t-")
		require.Equal(t, "\n\n", result.content)
	})

	t.Run("removes dashes from the text", func(t *testing.T) {
		result := parseContent("-\n- hello")
		require.Equal(t, "\n\nhello", result.content)
	})

	t.Run("turns second level bullet points into first level", func(t *testing.T) {
		result := parseContent("\t- hello\n\t- world")
		require.Equal(t, "\n- hello\n\n- world", result.content) // TODO: maybe remove the duplicated new line
	})

	t.Run("removes one tab from multi-level bullet points", func(t *testing.T) {
		result := parseContent("\t\t- hello\n\t\t\t- world")
		require.Equal(t, "\t- hello\n\t\t- world", result.content)
	})
	t.Run("removes tabs from all subsequent lines of a bullet point", func(t *testing.T) {
		result := parseContent(`
- ~~~ts
  const hello = "world";
  ~~~
- single line
- multiple
  lines
  in
  one`)
		require.Equal(t, `
~~~ts
const hello = "world";
~~~

single line

multiple
lines
in
one`, result.content)
	})
}
