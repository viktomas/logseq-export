package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
)

type page struct {
	filename   string
	attributes map[string]string
	text       string
}

/*
findMatchingFiles finds all files in rootPath that contain substring
ignoreRegexp is an expression that is evaluated on **relative** path of files within the graph (e.g. `.git/HEAD` or `logseq/.bkp/something.md`) if it matches, the file is not processed
*/
func findMatchingFiles(rootPath string, substring string, ignoreRegexp *regexp.Regexp) ([]string, error) {
	var result []string
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if d.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		if ignoreRegexp.MatchString(filepath.ToSlash(relativePath)) {
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
	unquotedProperties := flag.String("unquotedProperties", "", "comma-separated list of logseq page properties that won't be quoted in the markdown front matter, e.g. 'date,public,slug")
	flag.Parse()
	if *graphPath == "" || *blogFolder == "" {
		log.Println("mandatory argument is missing")
		flag.Usage()
		os.Exit(1)
	}
	publicFiles, err := findMatchingFiles(*graphPath, "public::", regexp.MustCompile(`^(logseq|.git)/`))
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
		result := transformPage(page)
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
