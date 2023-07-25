package main

import "regexp"

func parseTextAndAttributes(rawContent string) (string, map[string]string) {
	result := regexp.MustCompile(`^((?:.*?::.*\n)*)\n?((?:.|\s)+)$`).FindStringSubmatch(rawContent)
	attrArray := regexp.MustCompile(`(?m:^(.*?)::\s*(.*)$)`).FindAllStringSubmatch(result[1], -1)
	attributes := map[string]string{}
	for _, attrStrings := range attrArray {
		attributes[attrStrings[1]] = attrStrings[2]
	}
	return result[2], attributes
}

func parsePage(filename, rawContent string) oldPage {
	text, attributes := parseTextAndAttributes(rawContent)
	return oldPage{
		filename:   filename,
		attributes: attributes,
		text:       text,
	}
}
