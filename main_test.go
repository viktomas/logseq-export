package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// get path to the directory where this test file lives
var _, testFilepath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testFilepath)

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
		require.Equal(t, filepath.Join("/src", "pages", "b"), matchingFiles[0].absoluteFSPath)
		require.Equal(t, "public:: true\n- a bullet point", matchingFiles[0].content)
	})
}

func expectIdenticalContent(t testing.TB, expectedPath, actualPath string) {
	t.Helper()

	expectedContentBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatal("Error reading the expected file:", err)
	}
	actualContentBytes, err := os.ReadFile(actualPath)
	if err != nil {
		t.Fatal("Error reading the actual file:", err)
	}
	require.Equal(t, string(expectedContentBytes), string(actualContentBytes), fmt.Sprintf("the content of %s and %s are not identical", expectedPath, actualPath))
}

func makeRelative(t testing.TB, parent string, paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, absolutePath := range paths {
		relativePath, err := filepath.Rel(parent, absolutePath)
		if err != nil {
			t.Fatal("Can't produce relative path: ", err)
		}
		result = append(result, relativePath)
	}
	return result
}

var expectedAssets = []string{
	"logseq-assets/img-1.jpg",
	"logseq-assets/picture-2.png",
}

var expectedPages = []string{
	"logseq-pages/2023-07-29-not-so-complex.md",
	"logseq-pages/a.md",
	"logseq-pages/b.md",
}

func TestFullTransformation(t *testing.T) {
	deleteTestOutputFolder(t)
	testLogseqFolder := path.Join(testDir, "test", "logseq-folder")
	testOutputFolder := getTestOutputFolder(t)
	args := []string{
		"logseq-export",
		"--logseqFolder",
		testLogseqFolder,
		"--outputFolder",
		testOutputFolder,
	}
	err := Run(args)
	require.NoError(t, err)

	actualFiles := listFilesInFolder(t, testOutputFolder)

	t.Run("the files are moved correctly", func(t *testing.T) {
		require.Equal(
			t,
			append(expectedAssets, expectedPages...),
			makeRelative(t, testOutputFolder, actualFiles),
			"The list of files in output folder is different from what the test expected",
		)
	})

	t.Run("the files get transformed correctly", func(t *testing.T) {
		expectedOutputFolder := path.Join(testDir, "test", "expected-output")
		for i := 0; i < len(expectedPages); i++ {
			expectIdenticalContent(
				t,
				filepath.Join(expectedOutputFolder, expectedPages[i]),
				filepath.Join(testOutputFolder, expectedPages[i]),
			)
		}
	})
}

func TestRender(t *testing.T) {
	t.Run("it renders attributes as quoted strings", func(t *testing.T) {
		attributes := map[string]string{
			"first":  "1",
			"second": "2",
		}
		content := "page text"
		result := render(attributes, content, []string{})
		require.Equal(t, `---
first: "1"
second: "2"
---
page text`, result)
	})
	t.Run("it renders attributes in alphabetical order", func(t *testing.T) {
		attributes := map[string]string{
			"e": "1",
			"d": "1",
			"c": "1",
			"b": "1",
			"a": "1",
		}
		content := "page text"
		result := render(attributes, content, []string{})
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
		attributes := map[string]string{
			"first":  "1",
			"second": "2",
		}
		content := "page text"
		result := render(attributes, content, []string{"first", "second"})
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

	err := filepath.Walk(folderPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil

		}
		files = append(files, p)
		return nil
	})

	if err != nil {
		t.Fatalf("couldn't get all files in folder %v", folderPath)
	}

	sort.Strings(files)
	return files
}

func getTestOutputFolder(t testing.TB) string {
	t.Helper()
	return path.Join(testDir, "test", "test-output")
}

func deleteTestOutputFolder(t *testing.T) {
	t.Helper()
	// Specify the path of the folder you want to delete
	testOutputFolder := getTestOutputFolder(t)

	// Check if the folder exists
	if _, err := os.Stat(testOutputFolder); os.IsNotExist(err) {
		// The folder doesn't exist, nothing to delete
		return
	}

	// The folder exists, so delete it
	err := os.RemoveAll(testOutputFolder)
	if err != nil {
		t.Fatalf("Error deleting folder '%s': %s\n", testOutputFolder, err)
	}
}

func TestDetectPageLinks(t *testing.T) {
	content := `- created: 2021-02-28T11:04:46

TotT is a funny example of [[Environment design]] where Google decided to promote testing in 2006 by pasting one-page documents with tips and tricks on [[Automated testing]][^1]. It started as a joke during brainstorming session, but it turned out to be successful. Since 2006, there have been hundreds of episodes of one-page TotT.

[^1]: [[Winters, Manshreck, Wright - Software Engineering at Google]] p227
	`

	result := detectPageLinks(content)

	require.Equal(t, []string{"Environment design", "Automated testing", "Winters, Manshreck, Wright - Software Engineering at Google"}, result)
}
