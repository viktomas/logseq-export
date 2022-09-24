package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeName(t *testing.T) {
	result := sanitizeName("Blog idea%3A All good laws that EU brought.md")
	require.Equal(t, "Blog-idea%3A-All-good-laws-that-EU-brought.md", result)
}

func TestTransformPage(t *testing.T) {
	t.Run("sanitizes fileName", func(t *testing.T) {
		testPage := page{
			filename:   "name with space.md",
			attributes: map[string]string{},
			text:       "",
		}
		result := transformPage(testPage)
		require.Equal(t, "name-with-space.md", result.filename)
	})
}
