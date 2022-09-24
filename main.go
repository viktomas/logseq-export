package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
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
		srcContent, err := readFileToString(publicFile)
		if err != nil {
			log.Fatalf("Error when reading the %q file: %v", publicFile, err)
		}
		_, name := filepath.Split(publicFile)
		page := parsePage(name, srcContent)
		result := transformPage(page)
		dest := filepath.Join(exportFolder, result.filename)
		err = writeStringToFile(dest, render(result))
		if err != nil {
			log.Fatalf("Error when copying file %q: %v", dest, err)
		}
	}
}

func render(p page) string {
	attributeBuilder := strings.Builder{}
	for name, value := range p.attributes {
		attributeBuilder.WriteString(fmt.Sprintf("%s: %s\n", name, value))
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
