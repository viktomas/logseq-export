package main

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestLoadPublicPages(t *testing.T) {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	appFS.MkdirAll("/src/pages", 0755)
	appFS.MkdirAll("/src/logseq", 0755)
	afero.WriteFile(appFS, "/src/pages/b", []byte("public:: true\n- a bullet point"), 0644)
	afero.WriteFile(appFS, "/src/logseq/a", []byte("public:: true"), 0644)
	afero.WriteFile(appFS, "/src/pages/c", []byte("non public file"), 0644)

	t.Run("it finds files with 'public::' string in them", func(t *testing.T) {
		matchingFiles, err := loadPublicPages(appFS, "/src")

		require.Nil(t, err)
		require.Len(t, matchingFiles, 1)
		require.Equal(t, filepath.Join("/src", "pages", "b"), matchingFiles[0].fullPath)
		require.Equal(t, "public:: true\n- a bullet point", matchingFiles[0].content)
	})
}

func TestFullTransformation(t *testing.T) {
	testLogseqFolder := path.Join(path.Dir(t.Name()), "test/logseq-folder")
	testOutputFolder := path.Join(path.Dir(t.Name()), "test/test-output")
	args := []string{
		"logseq-export",
		"--logseqFolder",
		testLogseqFolder,
		"--outputFolder",
		testOutputFolder,
	}
	err := Run(args)
	require.NoError(t, err)

	expectedOutputFolder := path.Join(path.Dir(t.Name()), "test/expected-output")

	require.Equal(
		t,
		listFilesInFolder(t, expectedOutputFolder),
		listFilesInFolder(t, testOutputFolder),
		"The list of files in output folder is different from what the test expected",
	)
}

func TestRender(t *testing.T) {
	t.Run("it renders attributes as quoted strings", func(t *testing.T) {
		testPage := oldPage{
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
		testPage := oldPage{
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
		testPage := oldPage{
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

func listFilesInFolder(t *testing.T, folderPath string) []string {
	t.Helper()
	var files []string

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("couldn't get all files in folder %v", folderPath)
	}

	sort.Strings(files)
	return files
}
