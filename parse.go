package main

import (
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

func parsePage(publicPage textFile) parsedPage {
	pc := parseContent(publicPage.content)
	exportFilename := getExportFilename(publicPage.absoluteFSPath, pc.attributes)
	// add slug attribute if missing
	if _, ok := pc.attributes["slug"]; !ok {
		pc.attributes["slug"] = filenameWithoutExt(exportFilename)
	}
	// add title attribute if missing
	title, ok := pc.attributes["title"]
	if !ok {
		fileName := filepath.Base(publicPage.absoluteFSPath)
		title = getTitleFromFilename(fileName)
	}
	pc.attributes["title"] = title
	return parsedPage{
		exportFilename: exportFilename,
		originalPath:   publicPage.absoluteFSPath,
		pc:             pc,
	}
}

func filenameWithoutExt(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

func getTitleFromFilename(orig string) string {
	nameOnly := filenameWithoutExt(orig)
	unescaped, err := url.QueryUnescape(nameOnly)
	if err != nil {
		log.Printf("note name %q can't be unescaped because the %% sign is not followed by two hexadecimal characters", orig)
		unescaped = orig
	}
	return unescaped
}

func sanitizeName(orig string) string {
	ext := filepath.Ext(orig)
	title := getTitleFromFilename(orig)

	nonWordChars := regexp.MustCompile(`\W+`)
	sanitizedName := strings.ToLower(nonWordChars.ReplaceAllString(title, "-"))
	return strings.Join([]string{sanitizedName, ext}, "")
}

func getExportFilename(originalPath string, attributes map[string]string) string {
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

var dateLinkRegexp = regexp.MustCompile(`^\s*\[\[([^]]+?)]]\s*$`)

func parseAttributes(rawContent string) map[string]string {
	result := attrAndContentRegexp.FindStringSubmatch(rawContent)
	attrArray := regexp.MustCompile(`(?m:^(.*?)::\s*(.*)$)`).FindAllStringSubmatch(result[1], -1)
	attributes := map[string]string{}
	for _, attrStrings := range attrArray {
		attributes[attrStrings[1]] = attrStrings[2]
	}
	// remove link brackets from the date
	// [[2023-07-30]] -> 2023-07-30
	dateMatch := dateLinkRegexp.FindStringSubmatch(attributes["date"])
	if len(dateMatch) > 0 {
		attributes["date"] = dateMatch[1]
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

func removeTabFromMultiLevelBulletPoints(from string) string {
	return regexp.MustCompile(`(?m:^\t{1,}-)`).ReplaceAllStringFunc(from, func(s string) string {
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

func parseContent(rawContent string) parsedContent {
	content := applyStringTransformers(rawContent,
		stripAttributes,
		removeEmptyBulletPoints,
		unindentMultilineStrings,
		firstBulletPointsToParagraphs,
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
