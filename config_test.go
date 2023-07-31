package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseMandatoryFlags(t *testing.T) {
	args := []string{
		"script-name",
		"--logseqFolder",
		"/path/to/logseq",
		"--outputFolder",
		"/path/to/output",
	}

	config, err := parseConfig(args)

	if err != nil {
		t.Fatalf("error when parsing config: %v", err)
	}

	if config.LogseqFolder != "/path/to/logseq" {
		t.Fatalf("incorrectly parsed logseqFolder: %q", config.LogseqFolder)
	}

	if config.OutputFolder != "/path/to/output" {
		t.Fatalf("incorrectly parsed outputFolder. Expected %q got %q", "/path/to/logseq", config.OutputFolder)
	}

	if config.UnquotedProperties != nil {
		t.Fatalf("incorrectly parsed unquotedProperties. Expected nil, got %v", config.UnquotedProperties)
	}
}

func TestTestParsingOptionalFlags(t *testing.T) {
	configFolderPath := filepath.Join(filepath.Dir(t.Name()), "test/config")
	args := []string{
		"script-name",
		"--logseqFolder",
		configFolderPath,
		"--outputFolder",
		"/path/to/output",
	}

	config, err := parseConfig(args)

	if err != nil {
		t.Fatalf("error when parsing config: %v", err)
	}

	if config.LogseqFolder != configFolderPath {
		t.Fatalf("incorrectly parsed logseqFolder: %q", config.LogseqFolder)
	}

	if config.OutputFolder != "/path/to/output" {
		t.Fatalf("incorrectly parsed outputFolder. Expected %q got %q", configFolderPath, config.OutputFolder)
	}

	if !reflect.DeepEqual(config.UnquotedProperties, []string{"date", "tags"}) {
		t.Fatalf("incorrectly parsed unquotedProperties. Expected date, tags, got %v", config.UnquotedProperties)
	}
}
