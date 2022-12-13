package main

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestFindMatchingFiles(t *testing.T) {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	appFS.MkdirAll("/src/a", 0755)
	appFS.MkdirAll("/src/.git", 0755)
	appFS.MkdirAll("/src/logseq", 0755)
	afero.WriteFile(appFS, "/src/a/b", []byte("public:: true"), 0644)
	afero.WriteFile(appFS, "/src/.git/a", []byte("public:: true"), 0644)
	afero.WriteFile(appFS, "/src/logseq/a", []byte("public:: true"), 0644)
	afero.WriteFile(appFS, "/src/c", []byte("non public file"), 0644)

	t.Run("it finds files with 'public::' string in them", func(t *testing.T) {
		matchingFiles, err := findMatchingFiles(appFS, "/src", "public::", nil)

		require.Nil(t, err)
		require.Equal(t, []string{
			filepath.Join("/src", ".git", "a"),
			filepath.Join("/src", "a", "b"),
			filepath.Join("/src", "logseq", "a"),
		}, matchingFiles)
	})

	t.Run("it ignores files based on the ignoreRegex", func(t *testing.T) {
		matchingFiles, err := findMatchingFiles(appFS, "/src", "public::", regexp.MustCompile(`^(logseq|.git)/`))

		require.Nil(t, err)
		require.Equal(t, []string{filepath.Join("/src", "a", "b")}, matchingFiles)
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
		result := render(testPage, []string{}, []string{})
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
		result := render(testPage, []string{}, []string{})
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
		result := render(testPage, []string{"first", "second"}, []string{})
		require.Equal(t, `---
first: 1
second: 2
---
page text`, result)
	})
	t.Run("it renders formatted lists of attributes", func(t *testing.T) {
		testPage := page{
			filename: "",
			attributes: map[string]string{
				"first":  "single",
				"second": "tag1, tag2",
			},
			text: "page text",
		}
		result := render(testPage, []string{}, []string{"first", "second"})
		require.Equal(t, `---
first: ["single"]
second: ["tag1", "tag2"]
---
page text`, result)
	})
}
