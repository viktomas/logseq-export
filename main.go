package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
)

type page struct {
	filename   string
	attributes map[string]string
	assets     []string
	text       string
}

func findMatchingFiles(rootPath string, substring string) ([]string, error) {
	var result []string
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if d.IsDir() {
			return nil
		}
		file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
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
	graphPath := flag.String("graphPath", "", "[MANDATORY] Folder where all public pages are exported.")
	blogFolder := flag.String("blogFolder", "", "[MANDATORY] Folder where this program creates a new subfolder with public logseq pages.")
	assetsRelativePath := flag.String("assetsRelativePath", "", "relative path within blogFolder where the assets (images) should be stored (e.g. 'static/logseq')")
	imagePathPrefix := flag.String("imagePathPrefix", "", "path that the images are going to be served on on the web (e.g. '/public/images/logseq')")
	unquotedProperties := flag.String("unquotedProperties", "", "comma-separated list of logseq page properties that won't be quoted in the markdown front matter, e.g. 'date,public,slug")
	flag.Parse()
	if *graphPath == "" || *blogFolder == "" {
		log.Println("mandatory argument is missing")
		flag.Usage()
		os.Exit(1)
	}
	publicFiles, err := findMatchingFiles(*graphPath, "public::")
	if err != nil {
		log.Fatalf("Error during walking through a folder %v", err)
	}
	for _, publicFile := range publicFiles {
		srcContent, err := readFileToString(publicFile)
		if err != nil {
			log.Fatalf("Error when reading the %q file: %v", publicFile, err)
		}
		_, name := filepath.Split(publicFile)
		page := parsePage(name, srcContent)
		result := transformPage(page, *imagePathPrefix)
		assetFolder := filepath.Join(*blogFolder, *assetsRelativePath)
		err = os.MkdirAll(assetFolder, os.ModePerm)
		if err != nil {
			log.Fatalf("Error when making assets folder %q: %v", assetFolder, err)
		}
		err = copyAssets(publicFile, assetFolder, result.assets)
		if err != nil {
			log.Fatalf("Error when copying assets for page %q: %v", publicFile, err)
		}
		dest := filepath.Join(*blogFolder, result.filename)
		folder, _ := filepath.Split(dest)
		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			log.Fatalf("Error when creating parent directory for %q: %v", dest, err)
		}
		err = writeStringToFile(dest, render(result, parseUnquotedProperties(*unquotedProperties)))
		if err != nil {
			log.Fatalf("Error when copying file %q: %v", dest, err)
		}
	}
}

func copyAssets(baseFile string, assetFolder string, assets []string) error {
	baseDir, _ := filepath.Split(baseFile)
	for _, relativeAssetPath := range assets {
		assetPath := filepath.Clean(filepath.Join(baseDir, relativeAssetPath))
		_, assetName := filepath.Split(assetPath)
		destPath := filepath.Join(assetFolder, assetName)
		err := copy(assetPath, destPath)
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

func readFileToString(src string) (string, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()
	bytes, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func writeStringToFile(dest string, content string) error {
	err := os.WriteFile(dest, []byte(content), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func copy(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest) // creates if file doesn't exist
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}
	return nil
}
