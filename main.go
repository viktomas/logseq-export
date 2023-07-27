package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"golang.org/x/exp/slices"
)

type oldPage struct {
	filename   string
	attributes map[string]string
	assets     []string
	text       string
}

/*
findMatchingFiles finds all files in rootPath that contain substring
*/
func findMatchingFiles(appFS afero.Fs, rootPath string, substring string) ([]string, error) {
	var result []string
	err := afero.Walk(appFS, rootPath, func(path string, info fs.FileInfo, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if info.IsDir() {
			return nil
		}
		file, err := appFS.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		defer file.Close()
		fileScanner := bufio.NewScanner(file)
		for fileScanner.Scan() {
			line := fileScanner.Text()
			if strings.Contains(line, substring) {
				result = append(result, path)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type rawPage struct {
	fullPath string
	content  string
}

type page struct {
	originalPath string
	exportPath   string
	content      string
	attributes   map[string]string
}

func loadPublicPages(appFS afero.Fs, logseqFolder string) ([]rawPage, error) {
	logseqPagesFolder := path.Join(logseqFolder, "pages")
	publicFiles, err := findMatchingFiles(
		appFS,
		logseqPagesFolder,
		"public::",
	)
	// FIXME: test this error
	if err != nil {
		return nil, fmt.Errorf("error during walking through the logseq folder (%q): %w", logseqPagesFolder, err)
	}
	pages := make([]rawPage, 0, len(publicFiles))
	for _, publicFile := range publicFiles {
		srcContent, err := afero.ReadFile(appFS, publicFile)
		if err != nil {
			return nil, fmt.Errorf("reading the %q file failed: %w", publicFile, err)
		}
		pages = append(pages, rawPage{
			fullPath: publicFile,
			content:  string(srcContent),
		})
	}
	return pages, nil

}
func main() {
	err := Run(os.Args)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func Run(args []string) error {
	appFS := afero.NewOsFs()
	config, err := parseConfig(args)
	if err != nil {
		return fmt.Errorf("the configuration could not be parsed: %w", err)
	}
	publicPages, err := loadPublicPages(appFS, config.LogseqFolder)
	if err != nil {
		return fmt.Errorf("Error during walking through a folder %v", err)
	}

	exportPages := make([]page, 0, len(publicPages))
	for _, publicPage := range publicPages {
		exportPath := getExportPath(publicPage, config)
		exportPages = append(exportPages, page{
			originalPath: publicPage.fullPath,
			exportPath:   exportPath,
			content:      publicPage.content,
		})
	}

	for _, page := range exportPages {
		folder, _ := filepath.Split(page.exportPath)
		err = appFS.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating parent directory for %q failed: %v", page.exportPath, err)
		}
		err = afero.WriteFile(appFS, page.exportPath, []byte(page.content), 0644)
		if err != nil {
			return fmt.Errorf("copying file %q failed: %v", page.exportPath, err)
		}
	}
	return nil
}

func getExportPath(rawPage rawPage, config *Config) string {
	fileName := path.Base(rawPage.fullPath)
	return path.Join(config.OutputFolder, "logseq-pages", fileName)
}

func exportPublicPage(appFS afero.Fs, rawPage rawPage, config *Config) error {
	_, name := filepath.Split(rawPage.fullPath)
	page := parsePage(name, rawPage.content)
	result := transformPage(page, config.WebAssetsPathPrefix)
	assetFolder := filepath.Join(config.OutputFolder, config.AssetsRelativePath)
	err := copyAssets(appFS, rawPage.fullPath, assetFolder, result.assets)
	if err != nil {
		return fmt.Errorf("copying assets for page %q failed: %v", rawPage.fullPath, err)
	}
	dest := filepath.Join(config.OutputFolder, result.filename)
	folder, _ := filepath.Split(dest)
	err = appFS.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("creating parent directory for %q failed: %v", dest, err)
	}
	outputFileContent := render(result, config.UnquotedProperties)
	err = afero.WriteFile(appFS, dest, []byte(outputFileContent), 0644)
	if err != nil {
		return fmt.Errorf("copying file %q failed: %v", dest, err)
	}
	return nil
}

func copyAssets(appFS afero.Fs, baseFile string, assetFolder string, assets []string) error {
	err := appFS.MkdirAll(assetFolder, os.ModePerm)
	if err != nil {
		log.Fatalf("Error when making assets folder %q: %v", assetFolder, err)
	}
	baseDir, _ := filepath.Split(baseFile)
	for _, relativeAssetPath := range assets {
		assetPath := filepath.Clean(filepath.Join(baseDir, relativeAssetPath))
		_, assetName := filepath.Split(assetPath)
		destPath := filepath.Join(assetFolder, assetName)
		err := copy(appFS, assetPath, destPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseUnquotedProperties(param string) []string {
	if param == "" {
		return []string{}
	}
	return strings.Split(param, ",")
}

func render(p oldPage, dontQuote []string) string {
	sortedKeys := make([]string, 0, len(p.attributes))
	for k := range p.attributes {
		sortedKeys = append(sortedKeys, k)
	}
	slices.Sort(sortedKeys)
	attributeBuilder := strings.Builder{}
	for _, key := range sortedKeys {
		if slices.Contains(dontQuote, key) {
			attributeBuilder.WriteString(fmt.Sprintf("%s: %s\n", key, p.attributes[key]))
		} else {
			attributeBuilder.WriteString(fmt.Sprintf("%s: %q\n", key, p.attributes[key]))
		}
	}
	return fmt.Sprintf("---\n%s---\n%s", attributeBuilder.String(), p.text)
}
