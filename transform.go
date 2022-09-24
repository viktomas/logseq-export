package main

import (
	"fmt"
	"strings"
)

func sanitizeName(orig string) string {
	return strings.ReplaceAll(orig, " ", "-")
}

func generateFileName(originalName string, attributes map[string]string) string {
	if _, ok := attributes["slug"]; !ok {
		return sanitizeName(originalName)
	}
	if _, ok := attributes["date"]; ok {
		return fmt.Sprintf("%s-%s", attributes["date"], attributes["slug"])
	}

	return attributes["slug"]
}

func transformPage(p page) page {
	filename := generateFileName(p.filename, p.attributes)
	return page{
		filename:   filename,
		attributes: p.attributes,
		text:       p.text,
	}
}
