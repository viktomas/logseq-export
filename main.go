package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"golang.org/x/exp/slices"
)

type page struct {
	filename   string
	attributes map[string]string
	assets     []string
	text       string
}

/*
findMatchingFiles finds all files in rootPath that contain substring
ignoreRegexp is an expression that is evaluated on **relative** path of files within the graph (e.g. `.git/HEAD` or `logseq/.bkp/something.md`) if it matches, the file is not processed
*/
func findMatchingFiles(appFS afero.Fs, rootPath string, substring string, ignoreRegexp *regexp.Regexp) ([]string, error) {
	var result []string
	err := afero.Walk(appFS, rootPath, func(path string, info fs.FileInfo, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if info.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		if ignoreRegexp != nil && ignoreRegexp.MatchString(filepath.ToSlash(relativePath)) {
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

func main() {
	c, err := parseConfig(os.Args)
	if err != nil {
		log.Fatalf("parsing of the configuration failed: %v", err)
	}
	fmt.Printf("config is %v", c)
	appFS := afero.NewOsFs()
	config, err := parseConfig(os.Args)
	if err != nil {
		log.Fatalf("The configuration could not be parsed: %v", err)
	}
	publicFiles, err := findMatchingFiles(appFS, config.LogseqFolder, "public::", regexp.MustCompile(`^(logseq|.git)/`))
	if err != nil {
		log.Fatalf("Error during walking through a folder %v", err)
	}
	for _, publicFile := range publicFiles {
		err = exportPublicPage(appFS, publicFile, config)
		if err != nil {
			log.Fatalf("Error when exporting page %q: %v", publicFile, err)
		}
	}
}

func exportPublicPage(appFS afero.Fs, publicFile string, config *Config) error {
	srcContent, err := afero.ReadFile(appFS, publicFile)
	if err != nil {
		return fmt.Errorf("reading the %q file failed: %v", publicFile, err)
	}
	_, name := filepath.Split(publicFile)
	page := parsePage(name, string(srcContent))
	result := transformPage(page, config.WebAssetsPathPrefix)
	assetFolder := filepath.Join(config.OutputFolder, config.AssetsRelativePath)
	err = copyAssets(appFS, publicFile, assetFolder, result.assets)
	if err != nil {
		return fmt.Errorf("copying assets for page %q failed: %v", publicFile, err)
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

func render(p page, dontQuote []string) string {
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
