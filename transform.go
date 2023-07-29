package main

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

func sanitizeName(orig string) string {
	return strings.ReplaceAll(orig, " ", "-")
}

func generateFileName(originalName string, attributes map[string]string) string {
	slug, slugPresent := attributes["slug"]
	if !slugPresent {
		return sanitizeName(originalName)
	}

	if date, ok := attributes["date"]; ok {
		return fmt.Sprintf("%s.md", strings.Join(
			[]string{date, slug},
			"-",
		))
	}

	return fmt.Sprintf("%s.md", slug)
}

func addFileName(p oldPage) oldPage {
	filename := generateFileName(p.filename, p.attributes)
	folder := filepath.Join(path.Split(p.attributes["folder"])) // the page property always uses `/` but the final delimiter is OS-dependent
	p.filename = filepath.Join(folder, filename)
	return p
}

// onlyText turns text transformer into a page transformer
func onlyText(textTransformer func(string) string) func(oldPage) oldPage {
	return func(p oldPage) oldPage {
		p.text = textTransformer(p.text)
		return p
	}
}

func applyAll(from oldPage, transformers ...func(oldPage) oldPage) oldPage {
	result := from
	for _, t := range transformers {
		result = t(result)
	}
	return result
}

func blogAssetUrl(logseqURL, imagePrefixPath string) string {
	_, assetName := path.Split(logseqURL)
	return path.Join(imagePrefixPath, assetName)
}

func transformPage(p oldPage) oldPage {
	if p.attributes == nil {
		p.attributes = map[string]string{}
	}
	return applyAll(
		p,
		addFileName,
	)
}
