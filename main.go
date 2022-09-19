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
	"regexp"
	"strings"
	"time"
)

type page struct {
	filename   string
	attributes map[string]string
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

// Returns path to a newly created folder
func generateExportFolderName(outputFolder string) string {
	t := time.Now()
	timestamp := t.Format("2006-01-02-15-04-05")
	folderName := fmt.Sprintf("export-%s", timestamp)
	return filepath.Join(outputFolder, folderName)
}

func main() {
	graphPath := flag.String("graphPath", "", "[MANDATORY] Path to the root of your logseq graph containing /pages and /journals directories.")
	outputFolder := flag.String("outputFolder", "", "[MANDATORY] Folder where this program creates a new subfolder with public logseq pages.")
	flag.Parse()
	if *graphPath == "" || *outputFolder == "" {
		log.Println("mandatory argument is missing")
		flag.Usage()
		os.Exit(1)
	}
	publicFiles, err := findMatchingFiles(*graphPath, "public::")
	if err != nil {
		log.Fatalf("Error during walking through a folder %v", err)
	}
	exportFolder := generateExportFolderName(*outputFolder)
	err = os.Mkdir(exportFolder, os.ModePerm)
	if err != nil {
		log.Fatalf("Error when making the %q folder %v", exportFolder, err)
	}
	for _, publicFile := range publicFiles {
		log.Printf("copying %q", publicFile)
		_, name := filepath.Split(publicFile)
		dest := filepath.Join(exportFolder, sanitizeName(name))
		err = copyFile(publicFile, dest)
		if err != nil {
			log.Fatalf("Error when copying file %q: %v", dest, err)
		}
	}
}

func sanitizeName(orig string) string {
	return strings.ReplaceAll(orig, " ", "-")
}

func parseTextAndAttributes(rawContent string) (string, map[string]string) {
	result := regexp.MustCompile(`^((?:.*?::.*\n)*)\n?((?:.|\s)+)$`).FindStringSubmatch(rawContent)
	attrArray := regexp.MustCompile(`(?m:^(.*?)::\s*(.*)$)`).FindAllStringSubmatch(result[1], -1)
	attributes := map[string]string{}
	for _, attrStrings := range attrArray {
		attributes[attrStrings[1]] = attrStrings[2]
	}
	return result[2], attributes
}

func parsePage(filename, rawContent string) page {
	text, attributes := parseTextAndAttributes(rawContent)
	return page{
		filename:   filename,
		attributes: attributes,
		text:       text,
	}
}

func copyFile(src, dest string) error {

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}
	return nil
}
