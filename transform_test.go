package main

import (
	"path/filepath"
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
	testPage := oldPage{
		filename:   "",
		attributes: map[string]string{},
		text:       from,
	}
	result := transformPage(testPage, "")
	return result.text
}

// TODO make sure that all transform text ends with \n (for better diffs)
func TestTransformPage(t *testing.T) {
	t.Run("generates filename", func(t *testing.T) {
		testPage := oldPage{
			filename: "name with space.md",
			attributes: map[string]string{
				"slug": "this-is-a-slug",
				"date": "2022-09-24",
			},
			text: "",
		}
		result := transformPage(testPage, "")
		require.Equal(t, "2022-09-24-this-is-a-slug.md", result.filename)
	})

	t.Run("uses folder attribute in file name", func(t *testing.T) {
		testPage := oldPage{
			filename: "name with space.md",
			attributes: map[string]string{
				"folder": "content/posts",
			},
			text: "",
		}
		result := transformPage(testPage, "")
		require.Equal(t, filepath.Join("content", "posts", "name-with-space.md"), result.filename)
	})

}

func TestTransformImages(t *testing.T) {
	t.Run("extracts relative images", func(t *testing.T) {
		testPage := oldPage{
			filename: "a.md",
			text:     "- ![hello world](../assets/image.png)",
		}
		result := transformPage(testPage, "/images")
		require.Equal(t, []string{"../assets/image.png"}, result.assets)
		require.Equal(t, "- ![hello world](/images/image.png)", result.text)
	})

	t.Run("ignores absolute images", func(t *testing.T) {
		testPage := oldPage{
			filename: "a.md",
			text:     "- ![hello world](http://example.com/assets/image.png)",
		}
		result := transformPage(testPage, "/images")
		require.Equal(t, 0, len(result.assets))
		require.Equal(t, "- ![hello world](http://example.com/assets/image.png)", result.text)
	})

	t.Run("extracts relative images from image attribute", func(t *testing.T) {
		testPage := oldPage{
			attributes: map[string]string{
				"image": "../assets/image.png",
			},
			filename: "a.md",
			text:     "",
		}
		result := transformPage(testPage, "/images")
		require.Equal(t, []string{"../assets/image.png"}, result.assets)
		require.Equal(t, "/images/image.png", result.attributes["image"])
	})

	t.Run("ignores absolute images in image attribute", func(t *testing.T) {
		testPage := oldPage{
			attributes: map[string]string{
				"image": "http://example.com/assets/image.png",
			},
			filename: "a.md",
			text:     "",
		}
		result := transformPage(testPage, "/images")
		require.Equal(t, 0, len(result.assets))
		require.Equal(t, "http://example.com/assets/image.png", result.attributes["image"])
	})

}
