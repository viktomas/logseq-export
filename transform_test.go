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
		require.Equal(t, "2022-09-24-this-is-a-slug", result)
	})

	t.Run("combines slug and date into a filename", func(t *testing.T) {
		result := generateFileName("name with space.md", map[string]string{
			"slug": "this-is-a-slug",
			"date": "2022-09-24",
		})
		require.Equal(t, "2022-09-24-this-is-a-slug", result)
	})

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
		require.Equal(t, "2022-09-24-this-is-a-slug", result.filename)
	})
}
