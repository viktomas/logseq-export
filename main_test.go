package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	t.Run("it renders attributes as quoted strings", func(t *testing.T) {
		testPage := page{
			filename: "",
			attributes: map[string]string{
				"first":  "1",
				"second": "2",
			},
			text: "page text",
		}
		result := render(testPage, []string{})
		require.Equal(t, `---
first: "1"
second: "2"
---
page text`, result)
	})
	t.Run("it renders attributes without quotes", func(t *testing.T) {
		testPage := page{
			filename: "",
			attributes: map[string]string{
				"first":  "1",
				"second": "2",
			},
			text: "page text",
		}
		result := render(testPage, []string{"first", "second"})
		require.Equal(t, `---
first: 1
second: 2
---
page text`, result)
	})
}
