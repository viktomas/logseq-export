package main

import "regexp"

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

func parsePage(filename, rawContent string) oldPage {
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
