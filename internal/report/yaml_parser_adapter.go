package internal

import (
	"io"
	"strings"
)

// YAMLParserAdapter parses YAML files as plain text so any clone finder can
// analyze .yml/.yaml documents through the common framework pipeline.
type YAMLParserAdapter struct{}

func (y *YAMLParserAdapter) Name() string { return "yaml" }

func (y *YAMLParserAdapter) SupportedFormats() []string { return []string{".yaml", ".yml"} }

func (y *YAMLParserAdapter) Parse(reader io.Reader) ([]string, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return []string{strings.TrimSpace(string(b))}, nil
}
