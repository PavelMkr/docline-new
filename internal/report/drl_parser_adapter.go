package internal

import (
	"encoding/xml"
	"io"
	"strings"
)

// DRLParserAdapter extracts *all* text nodes from a DRL XML document.
// This ensures clone analysis considers text inside <d:InfElement> and other elements.
type DRLParserAdapter struct{}

func (d *DRLParserAdapter) Name() string { return "drl" }

func (d *DRLParserAdapter) SupportedFormats() []string { return []string{".drl"} }

func (d *DRLParserAdapter) Parse(reader io.Reader) ([]string, error) {
	dec := xml.NewDecoder(reader)
	dec.Strict = false // be tolerant to slightly invalid XML

	var blocks []string
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch t := tok.(type) {
		case xml.CharData:
			s := strings.TrimSpace(string(t))
			if s != "" {
				blocks = append(blocks, s)
			}
		}
	}

	// One segment is enough; the framework joins segments with '\n' anyway.
	return []string{strings.Join(blocks, "\n")}, nil
}