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
	})
}

const multilineBlocks = `\n?(- .*\n(?:  .*\n?)+)`

/*
Makes sure that code blocks and multiline blocks are without any extra characters at the start of the line

  - ```ts
    const hello = "world"
    ```

is changed to

```ts
const hello = "world"
```
*/
func unindentMultilineStrings(from string) string {
	return regexp.MustCompile(multilineBlocks).ReplaceAllStringFunc(from, func(s string) string {
		match := regexp.MustCompile(multilineBlocks).FindStringSubmatch(s)
		onlyBlock := match[1]
		replacement := regexp.MustCompile(`((?m:^[- ] ))`).ReplaceAllString(onlyBlock, "") // remove the leading spaces or dash
		replacedString := strings.Replace(s, onlyBlock, replacement, 1)
		return fmt.Sprintf("\n%s", replacedString) // add extra new line
	})
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
		unindentMultilineStrings,
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
