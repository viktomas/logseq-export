package main

import (
	"fmt"
	"regexp"
	"strings"
)

func sanitizeName(orig string) string {
	return strings.ReplaceAll(orig, " ", "-")
}

func generateFileName(originalName string, attributes map[string]string) string {
	if _, ok := attributes["slug"]; !ok {
		return sanitizeName(originalName)
	}

	var date string
	if _, ok := attributes["date"]; ok {
		date = fmt.Sprintf("%s-", attributes["date"])
	}

	return fmt.Sprintf("%s%s.md", date, attributes["slug"])
}

func removeEmptyBulletPoints(from string) string {
	return regexp.MustCompile(`(?m:^\s*-\s*$)`).ReplaceAllString(from, "")
}

func firstBulletPointsToParagraphs(from string) string {
	return regexp.MustCompile(`(?m:^- )`).ReplaceAllString(from, "\n")
}

func secondToFirstBulletPoints(from string) string {
	return regexp.MustCompile(`(?m:^\t-)`).ReplaceAllString(from, "\n-")
}

func removeTabFromMultiLevelBulletPoints(from string) string {
	return regexp.MustCompile(`(?m:^\t{2,}-)`).ReplaceAllStringFunc(from, func(s string) string {
		return s[1:]
	}) //TODO make general (it can be merged with previous rule)
}

func applyAll(from string, transformers ...func(string) string) string {
	result := from
	for _, t := range transformers {
		result = t(result)
	}
	return result
}

func transformPage(p page) page {
	filename := generateFileName(p.filename, p.attributes)
	text := applyAll(
		p.text,
		removeEmptyBulletPoints,
		firstBulletPointsToParagraphs,
		secondToFirstBulletPoints,
		removeTabFromMultiLevelBulletPoints,
	)
	return page{
		filename:   filename,
		attributes: p.attributes,
		text:       text,
	}
}
