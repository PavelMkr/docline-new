package internal

import (
	"fmt"
	"io"

	"Docline/internal/framework"
)

// DocBookParserAdapter adapts DocBookParser to the framework.DocumentParser interface.
type DocBookParserAdapter struct{}

func (d *DocBookParserAdapter) Name() string {
	return "docbook"
}

func (d *DocBookParserAdapter) SupportedFormats() []string {
	return []string{".xml", ".dbk", ".docbook"}
}

func (d *DocBookParserAdapter) Parse(reader io.Reader) ([]string, error) {
	parser := NewDocBookParser()
	return parser.ParseDocBook(reader)
}

// PandocConverterAdapter adapts DocumentConverter to the framework.DocumentConverter interface.
type PandocConverterAdapter struct {
	converter *DocumentConverter
}

func NewPandocConverterAdapter() *PandocConverterAdapter {
	return &PandocConverterAdapter{
		converter: NewDocumentConverter(),
	}
}

func (p *PandocConverterAdapter) Name() string {
	return "pandoc"
}

func (p *PandocConverterAdapter) Convert(inputPath string, outputFormat string) (string, error) {
	if !p.isSupportedOutput(outputFormat) {
		return "", fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	if p.converter == nil {
		p.converter = NewDocumentConverter()
	}
	// Our underlying converter always produces DocBook XML; the outputFormat
	// is used only for validation at this layer.
	return p.converter.ConvertToDocBook(inputPath)
}

func (p *PandocConverterAdapter) IsConversionNeeded(filePath string) bool {
	if p.converter == nil {
		p.converter = NewDocumentConverter()
	}
	return p.converter.IsConversionNeeded(filePath)
}

func (p *PandocConverterAdapter) SupportedInputFormats() []string {
	if p.converter == nil {
		p.converter = NewDocumentConverter()
	}
	var exts []string
	for ext := range p.converter.SupportedInputFormats {
		exts = append(exts, ext)
	}
	return exts
}

func (p *PandocConverterAdapter) SupportedOutputFormats() []string {
	if p.converter == nil {
		p.converter = NewDocumentConverter()
	}
	var exts []string
	for ext := range p.converter.SupportedOutputFormats {
		exts = append(exts, ext)
	}
	return exts
}

func (p *PandocConverterAdapter) isSupportedOutput(ext string) bool {
	if p.converter == nil {
		p.converter = NewDocumentConverter()
	}
	return p.converter.SupportedOutputFormats[ext]
}

// RegisterDocumentPlugins registers the built-in parser and converter in the plugin registry.
func RegisterDocumentPlugins(reg *framework.PluginRegistry) error {
	if err := reg.RegisterDocumentParser(&DocBookParserAdapter{}); err != nil {
		return fmt.Errorf("register docbook parser: %w", err)
	}
	if err := reg.RegisterDocumentConverter(NewPandocConverterAdapter()); err != nil {
		return fmt.Errorf("register pandoc converter: %w", err)
	}
	return nil
}
