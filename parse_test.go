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
}

func TestParsePage(t *testing.T) {
	t.Run("adds filename as title if it is missing", func(t *testing.T) {
		testPage := textFile{
			absoluteFSPath: "/name with space.md",
			content:        "",
		}
		result := parsePage(testPage)
		require.Equal(t, "name with space", result.pc.attributes["title"])
	})

	t.Run("uses title page property if present", func(t *testing.T) {
		testPage := textFile{
			absoluteFSPath: "/name with space.md",
			content:        "title:: title from page prop\n",
		}
		result := parsePage(testPage)
		require.Equal(t, "title from page prop", result.pc.attributes["title"])
	})

	t.Run("uses sanitized filename as the exportFileName", func(t *testing.T) {
		testPage := textFile{
			absoluteFSPath: "/Blog idea%3A All good laws that EU brought.md",
			content:        "",
		}
		result := parsePage(testPage)
		require.Equal(t, "Blog-idea%3A-All-good-laws-that-EU-brought.md", result.exportFilename)
	})

	t.Run("uses slug as the exportFileName", func(t *testing.T) {
		testPage := textFile{
			absoluteFSPath: "/name with space.md",
			content:        "slug:: slug-name\n",
		}
		result := parsePage(testPage)
		require.Equal(t, "slug-name.md", result.exportFilename)
	})

	t.Run("uses date and slug as the exportFileName", func(t *testing.T) {
		testPage := textFile{
			absoluteFSPath: "/name with space.md",
			content:        "slug:: slug-name\ndate:: 2023-07-29\n",
		}
		result := parsePage(testPage)
		require.Equal(t, "2023-07-29-slug-name.md", result.exportFilename)
	})

	t.Run("keeps slug if present", func(t *testing.T) {
		testPage := textFile{
			absoluteFSPath: "/name with space.md",
			content:        "slug:: slug-name\n",
		}
		result := parsePage(testPage)
		require.Equal(t, "slug-name", result.pc.attributes["slug"])
	})

	t.Run("uses exportFilename as slug", func(t *testing.T) {
		testPage := textFile{
			absoluteFSPath: "/name with space.md",
			content:        "",
		}
		result := parsePage(testPage)
		require.Equal(t, "name-with-space", result.pc.attributes["slug"])
	})
}

func TestParseContent(t *testing.T) {
	t.Run("parses page with only one attribute", func(t *testing.T) {
		result := parseContent("public:: true\n")
		require.Equal(t, "", result.content)
		require.Equal(t, "true", result.attributes["public"])
	})

	t.Run("parses page with one line", func(t *testing.T) {
		result := parseContent("- a\n")
		require.Equal(t, "\na\n", result.content)
		require.Empty(t, result.attributes)
	})

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
