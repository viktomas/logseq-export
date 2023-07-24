package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"path"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	LogseqFolder        string
	OutputFolder        string
	UnquotedProperties  []string
	AssetsRelativePath  string
	WebAssetsPathPrefix string
}

func (c *Config) Validate() error {
	if c.LogseqFolder == "" {
		return errors.New("logseqFolder command line argument is mandatory ")
	}
	if c.OutputFolder == "" {
		return errors.New("outputFolder command line argument is mandatory ")
	}
	return nil
}

func flagset() *flag.FlagSet {
	f := flag.NewFlagSet("config", flag.ExitOnError)
	f.String("logseqFolder", "", "[MANDATORY] Folder where all public pages are exported.")
	f.String("outputFolder", "", "[MANDATORY] Folder where the transformed logseq pages will be stored.")
	return f
}

func parseConfig(args []string) (*Config, error) {
	var config Config

	f := flagset()

	k := koanf.New(".")

	if args == nil {
		args = []string{}
	} else if len(args) > 0 {
		args = args[1:]
	}

	// Parse args
	if err := f.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	// Load parsed args to koanf
	if err := k.Load(basicflag.Provider(f, "."), nil); err != nil {
		return nil, fmt.Errorf("error loading default config from flags: %w", err)
	}

	logseqFolder := k.String("logseqFolder")
	// Load YAML config and merge into the previously loaded config (because we can).
	configPath := path.Join(logseqFolder, "export.yaml")
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		log.Printf("Failed to read config file %q. Using default config.", configPath)
	}

	if err := k.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("error unmarshal config: %w", err)
	}

	err := config.Validate()
	if err != nil {
		return nil, err
	}
	return &config, nil
}
