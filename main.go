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
)

func findMatchingFiles(rootPath string, substring string) ([]string, error) {
	var result []string
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, walkError error) error {
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
	graphPath := flag.String("graphPath", "", "Path to the root of your logseq graph containing /pages and /journals directories.")
	flag.Parse()
	if *graphPath == "" {
		log.Println("graphPath argument is mandatory")
		flag.Usage()
		os.Exit(1)
	}
	result, err := findMatchingFiles(*graphPath, "public::")
	if err != nil {
		log.Fatalf("Error during walking through a folder %v", err)
	}
	fmt.Println(result)
}
