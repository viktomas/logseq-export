package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

func parsePage(publicPage textFile) parsedPage {
	pc := parseContent(publicPage.content)
	exportFilename := generateFileName(publicPage.absoluteFSPath, pc.attributes)
	ensureSlugInAttributes(pc, exportFilename)
	title := parseTitle(pc, publicPage.absoluteFSPath)
	pc.attributes["title"] = title
	return parsedPage{
		exportFilename: exportFilename,
		originalPath:   publicPage.absoluteFSPath,
		pc:             pc,
	}
}

func ensureSlugInAttributes(pc parsedContent, exportFilename string) {
	if _, ok := pc.attributes["slug"]; !ok {
		pc.attributes["slug"] = exportFilename[:len(exportFilename)-len(filepath.Ext(exportFilename))]
	}
}

func sanitizeName(orig string) string {
	return strings.ReplaceAll(orig, " ", "-")
}

func generateFileName(originalPath string, attributes map[string]string) string {
	originalName := filepath.Base(originalPath)
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

var attrAndContentRegexp = regexp.MustCompile(`^((?:.*?::.*\n)*)\n?((?:.|\s)+)?$`)

func parseAttributes(rawContent string) map[string]string {
	result := attrAndContentRegexp.FindStringSubmatch(rawContent)
	attrArray := regexp.MustCompile(`(?m:^(.*?)::\s*(.*)$)`).FindAllStringSubmatch(result[1], -1)
	attributes := map[string]string{}
	for _, attrStrings := range attrArray {
		attributes[attrStrings[1]] = attrStrings[2]
	}
	return attributes
}

func stripAttributes(rawContent string) string {
	result := attrAndContentRegexp.FindStringSubmatch(rawContent)
	return result[2]
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

func applyStringTransformers(from string, transformers ...func(string) string) string {
	result := from
	for _, t := range transformers {
		result = t(result)
	}
	return result
}

func parseTitle(pc parsedContent, absoluteFSPath string) string {
	title, ok := pc.attributes["title"]
	fileName := filepath.Base(absoluteFSPath)
	if !ok {
		title = regexp.MustCompile(`\.[^.]*$`).ReplaceAllString(fileName, "")
	}
	return title
}

func parseContent(rawContent string) parsedContent {
	content := applyStringTransformers(rawContent,
		stripAttributes,
		removeEmptyBulletPoints,
		unindentMultilineStrings,
		firstBulletPointsToParagraphs,
		secondToFirstBulletPoints,
		removeTabFromMultiLevelBulletPoints,
	)
	return parsedContent{
		attributes: parseAttributes(rawContent),
		content:    content,
		assets:     parseAssets(rawContent),
	}
}

/*
parseAssets finds all paths to relative markdown images
![img](../assets/img.jpg) - returns `../assets/img.jpg`
![img](http://example.com/img/jpg) - is ignored
*/
func parseAssets(content string) []string {
	assetRegexp := regexp.MustCompile(`!\[.*?]\((\.\.?/.+?)\)`)
	links := assetRegexp.FindAllStringSubmatch(content, -1)
	assets := make([]string, 0, len(links))
	for _, l := range links {
		assets = append(assets, l[1])
	}
	return assets
}
