package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeName(t *testing.T) {
	result := sanitizeName("Blog idea%3A All good laws that EU brought.md")
	require.Equal(t, "Blog-idea%3A-All-good-laws-that-EU-brought.md", result)
}

func GenerateFileName(t *testing.T) {

	t.Run("sanitizes fileName if there's no slug", func(t *testing.T) {
		result := generateFileName("name with space.md", map[string]string{})
		require.Equal(t, "name-with-space.md", result)
	})

	t.Run("combines slug and date into a filename", func(t *testing.T) {
		result := generateFileName("name with space.md", map[string]string{
			"slug": "this-is-a-slug",
			"date": "2022-09-24",
		})
		require.Equal(t, "2022-09-24-this-is-a-slug.md", result)
	})

	t.Run("combines slug and date into a filename", func(t *testing.T) {
		result := generateFileName("name with space.md", map[string]string{
			"slug": "this-is-a-slug",
			"date": "2022-09-24",
		})
		require.Equal(t, "2022-09-24-this-is-a-slug.md", result)
	})

}

func transformText(from string) string {
	testPage := page{
		filename:   "",
		attributes: map[string]string{},
		text:       from,
	}
	result := transformPage(testPage)
	return result.text
}

func TestTransformPage(t *testing.T) {
	t.Run("generates filename", func(t *testing.T) {
		testPage := page{
			filename: "name with space.md",
			attributes: map[string]string{
				"slug": "this-is-a-slug",
				"date": "2022-09-24",
			},
			text: "",
		}
		result := transformPage(testPage)
		require.Equal(t, "2022-09-24-this-is-a-slug.md", result.filename)
	})

	t.Run("uses folder attribute in file name", func(t *testing.T) {
		testPage := page{
			filename: "name with space.md",
			attributes: map[string]string{
				"folder": "posts",
			},
			text: "",
		}
		result := transformPage(testPage)
		require.Equal(t, "posts/name-with-space.md", result.filename)
	})

	t.Run("removes dashes with no text after them", func(t *testing.T) {
		result := transformText("-\n\t- \n\t\t-")
		require.Equal(t, "\n\n", result)
	})

	t.Run("removes dashes from the text", func(t *testing.T) {
		result := transformText("-\n- hello")
		require.Equal(t, "\n\nhello", result)
	})

	t.Run("turns second level bullet points into first level", func(t *testing.T) {
		result := transformText("\t- hello\n\t- world")
		require.Equal(t, "\n- hello\n\n- world", result) // TODO: maybe remove the duplicated new line
	})

	t.Run("removes one tab from multi-level bullet points", func(t *testing.T) {
		result := transformText("\t\t- hello\n\t\t\t- world")
		require.Equal(t, "\t- hello\n\t\t- world", result)
	})
	t.Run("removes tabs from all subsequent lines of a bullet point", func(t *testing.T) {
		result := transformText(`
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
one`, result)
	})
}
