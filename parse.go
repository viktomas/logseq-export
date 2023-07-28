package main

import (
	"fmt"
	"regexp"
	"strings"
)

func parseAttributes(rawContent string) map[string]string {
	result := regexp.MustCompile(`^((?:.*?::.*\n)*)\n?((?:.|\s)+)$`).FindStringSubmatch(rawContent)
	attrArray := regexp.MustCompile(`(?m:^(.*?)::\s*(.*)$)`).FindAllStringSubmatch(result[1], -1)
	attributes := map[string]string{}
	for _, attrStrings := range attrArray {
		attributes[attrStrings[1]] = attrStrings[2]
	}
	return attributes
}

func stripAttributes(rawContent string) string {
	result := regexp.MustCompile(`^((?:.*?::.*\n)*)\n?((?:.|\s)+)$`).FindStringSubmatch(rawContent)
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

/* Deprecated: use the parseContent instead */
func parsePageOld(filename, rawContent string) oldPage {
	return oldPage{
		filename:   filename,
		attributes: parseAttributes(rawContent),
		text:       stripAttributes(rawContent),
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
