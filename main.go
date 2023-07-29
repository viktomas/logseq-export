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

/* textFile captures all data about a text file stored on disk that we need for exporting logseq graph */
type textFile struct {
	absoluteFSPath string
	content        string
}

type parsedContent struct {
	/* content without attributes */
	content    string
	attributes map[string]string
	assets     []string
}

type parsedPage struct {
	title          string
	exportFilename string
	originalPath   string
	pc             parsedContent
}

const publicAttributeSubstring = "public::"

func loadPublicPages(appFS afero.Fs, logseqFolder string) ([]textFile, error) {
	logseqPagesFolder := path.Join(logseqFolder, "pages")
	// Find all files that contain `public::`
	var publicFiles []string
	err := afero.Walk(appFS, logseqPagesFolder, func(path string, info fs.FileInfo, walkError error) error {
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
			if strings.Contains(line, publicAttributeSubstring) {
				publicFiles = append(publicFiles, path)
				return nil
			}
		}
		return nil
	})
	// FIXME: test this error
	if err != nil {
		return nil, fmt.Errorf("error during walking through the logseq folder (%q): %w", logseqPagesFolder, err)
	}
	pages := make([]textFile, 0, len(publicFiles))
	for _, publicFile := range publicFiles {
		srcContent, err := afero.ReadFile(appFS, publicFile)
		if err != nil {
			return nil, fmt.Errorf("reading the %q file failed: %w", publicFile, err)
		}
		pages = append(pages, textFile{
			absoluteFSPath: publicFile,
			content:        string(srcContent),
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

	// parse pages
	exportPages := make([]parsedPage, 0, len(publicPages))
	for _, publicPage := range publicPages {
		exportPages = append(exportPages, parsePage(publicPage))
	}

	err = exportAssets(appFS, config.OutputFolder, exportPages)
	if err != nil {
		return fmt.Errorf("failed to export assets: %w", err)
	}

	for _, page := range exportPages {
		exportPath := getExportPath(page.exportFilename, config)
		folder, _ := filepath.Split(exportPath)
		err = appFS.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating parent directory for %q failed: %v", exportPath, err)
		}
		page.pc.attributes["title"] = page.title
		// TODO: more processing on the content (linking pages, attributes)
		finalContent := replaceAssetPaths(page)
		// TODO find out what properties should I not quote
		err = afero.WriteFile(appFS, exportPath, []byte(render(page.pc.attributes, finalContent, []string{"public"})), 0644)
		if err != nil {
			return fmt.Errorf("copying file %q failed: %v", exportPath, err)
		}
	}
	return nil
}

func exportAssets(appFS afero.Fs, outputFolder string, exportPages []parsedPage) error {
	// get all asset paths (deduplicated)
	assetFullPaths := map[string]struct{}{}
	for _, page := range exportPages {
		for _, assetPath := range page.pc.assets {
			fullPath := filepath.Clean(filepath.Join(filepath.Dir(page.originalPath), assetPath))
			assetFullPaths[fullPath] = struct{}{}
		}
	}

	assetOutputFolder := filepath.Join(outputFolder, "logseq-assets")

	assetSrcAndDest := map[string]string{}
	for fullPath := range assetFullPaths {
		dest := filepath.Join(assetOutputFolder, filepath.Base(fullPath))
		assetSrcAndDest[fullPath] = dest
	}

	err := appFS.MkdirAll(assetOutputFolder, os.ModePerm)
	if err != nil {
		log.Fatalf("Error when making assets folder %q: %v", assetOutputFolder, err)
	}

	for src, dest := range assetSrcAndDest {
		err = copy(appFS, src, dest)
		if err != nil {
			return fmt.Errorf("failed copying asset from %q to %q: %w", src, dest, err)
		}
	}
	return nil
}

func replaceAssetPaths(p parsedPage) string {
	newContent := p.pc.content
	for _, link := range p.pc.assets {
		fileName := filepath.Base(link)
		newContent = strings.ReplaceAll(newContent, link, path.Join("/logseq-assets", fileName))
	}
	return newContent
}

func getExportPath(filename string, config *Config) string {
	return path.Join(config.OutputFolder, "logseq-pages", filename)
}

func parseUnquotedProperties(param string) []string {
	if param == "" {
		return []string{}
	}
	return strings.Split(param, ",")
}

func render(attributes map[string]string, content string, dontQuote []string) string {
	sortedKeys := make([]string, 0, len(attributes))
	for k := range attributes {
		sortedKeys = append(sortedKeys, k)
	}
	slices.Sort(sortedKeys)
	attributeBuilder := strings.Builder{}
	for _, key := range sortedKeys {
		if slices.Contains(dontQuote, key) {
			attributeBuilder.WriteString(fmt.Sprintf("%s: %s\n", key, attributes[key]))
		} else {
			attributeBuilder.WriteString(fmt.Sprintf("%s: %q\n", key, attributes[key]))
		}
	}
	return fmt.Sprintf("---\n%s---\n%s", attributeBuilder.String(), content)
}
