package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindMatchingFiles(t *testing.T) {
	t.Run("it finds files with 'public::' string in them", func(t *testing.T) {

	})
}

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
	t.Run("it renders attributes in alphabetical order", func(t *testing.T) {
		testPage := page{
			filename: "",
			attributes: map[string]string{
				"e": "1",
				"d": "1",
				"c": "1",
				"b": "1",
				"a": "1",
			},
			text: "page text",
		}
		result := render(testPage, []string{})
		require.Equal(t, `---
a: "1"
b: "1"
c: "1"
d: "1"
e: "1"
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
