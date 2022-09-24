package main

import "strings"

func sanitizeName(orig string) string {
	return strings.ReplaceAll(orig, " ", "-")
}

func generateFileName(originalName string, attributes map[string]string) string {
	return sanitizeName(originalName)
}

func transformPage(p page) page {
	filename := generateFileName(p.filename, p.attributes)
	return page{
		filename:   filename,
		attributes: p.attributes,
		text:       p.text,
	}
}
