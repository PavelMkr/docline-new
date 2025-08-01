package main

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDocBookParser(t *testing.T) {
	parser := NewDocBookParser()

	// check that all standard elements are present
	expectedElements := []string{"para", "section", "chapter", "title", "simpara", "note", "warning", "important", "tip"}
	for _, elem := range expectedElements {
		if !parser.TextElements[elem] {
			t.Errorf("Expected element %s to be in TextElements, but it wasn't", elem)
		}
	}
}

func TestDocBookParser_AddRemoveTextElement(t *testing.T) {
	parser := NewDocBookParser()

	// test adding a new element
	newElement := "custom_element"
	parser.AddTextElement(newElement)
	if !parser.TextElements[newElement] {
		t.Errorf("Failed to add new text element %s", newElement)
	}

	// test removing an element
	parser.RemoveTextElement(newElement)
	if parser.TextElements[newElement] {
		t.Errorf("Failed to remove text element %s", newElement)
	}
}

func TestDocBookParser_ParseDocBook(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{
			name: "Valid DocBook with para elements",
			input: `<?xml version="1.0" encoding="UTF-8"?>
				<book>
					<para>First paragraph</para>
					<para>Second paragraph</para>
				</book>`,
			expected: []string{"First paragraph", "Second paragraph"},
			wantErr:  false,
		},
		{
			name: "Valid DocBook with nested elements",
			input: `<?xml version="1.0" encoding="UTF-8"?>
				<book>
					<chapter>
						<title>Chapter Title</title>
						<para>Chapter content</para>
					</chapter>
				</book>`,
			expected: []string{"Chapter Title", "Chapter content"},
			wantErr:  false,
		},
		{
			name: "DocBook with HTML entities",
			input: `<?xml version="1.0" encoding="UTF-8"?>
				<book>
					<para>Text with &amp; and &lt;entities&gt;</para>
				</book>`,
			expected: []string{"Text with & and <entities>"},
			wantErr:  false,
		},
		{
			name:    "Invalid XML",
			input:   `<invalid>xml</invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewDocBookParser()
			reader := strings.NewReader(tt.input)

			segments, err := parser.ParseDocBook(reader)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(segments) != len(tt.expected) {
				t.Errorf("Expected %d segments, got %d", len(tt.expected), len(segments))
				return
			}

			for i, expected := range tt.expected {
				if segments[i] != expected {
					t.Errorf("Segment %d: expected %q, got %q", i, expected, segments[i])
				}
			}
		})
	}
}

func TestDocBookParser_ExtractText(t *testing.T) {
	parser := NewDocBookParser()

	// create test DocBook structure
	doc := &DocBookElement{
		XMLName: xml.Name{Local: "book"},
		Elements: []DocBookElement{
			{
				XMLName: xml.Name{Local: "para"},
				Content: "Test content",
			},
			{
				XMLName: xml.Name{Local: "section"},
				Elements: []DocBookElement{
					{
						XMLName: xml.Name{Local: "para"},
						Content: "Nested content",
					},
				},
			},
		},
	}

	var segments []string
	parser.extractText(doc, &segments)

	expected := []string{"Test content", "Nested content"}
	if len(segments) != len(expected) {
		t.Errorf("Expected %d segments, got %d", len(expected), len(segments))
		return
	}

	for i, exp := range expected {
		if segments[i] != exp {
			t.Errorf("Segment %d: expected %q, got %q", i, exp, segments[i])
		}
	}
}

func TestDocBookParser_RealDocBookFile(t *testing.T) {
	// path to DocBook in test folder
	docbookFile := filepath.Join("test", "DocBook_Definitive_Guide.xml")

	// check that file exists
	if _, err := os.Stat(docbookFile); os.IsNotExist(err) {
		t.Skipf("DocBook file not found: %s", docbookFile)
		return
	}

	// open file
	file, err := os.Open(docbookFile)
	if err != nil {
		t.Fatalf("Failed to open DocBook file: %v", err)
	}
	defer file.Close()

	// create parser and parse file
	parser := NewDocBookParser()
	segments, err := parser.ParseDocBook(file)

	if err != nil {
		t.Fatalf("Failed to parse DocBook file: %v", err)
	}

	// check that we got at least some segments
	if len(segments) == 0 {
		t.Error("Expected at least one text segment from DocBook file, got none")
		return
	}

	// check that segments are not empty
	for i, segment := range segments {
		if strings.TrimSpace(segment) == "" {
			t.Errorf("Segment %d is empty or contains only whitespace", i)
		}
	}

	t.Logf("Successfully parsed DocBook file, found %d text segments", len(segments))

	// print first few segments for debugging
	for i, segment := range segments {
		if i >= 3 { // print only first 3 segments
			break
		}
		t.Logf("Segment %d: %s", i, strings.TrimSpace(segment[:min(len(segment), 100)]))
	}
}

func TestDocBookParser_LinuxKernelDoc(t *testing.T) {
	// path to Linux Kernel documentation file
	kernelDocFile := filepath.Join("test", "Linux_Kernel_Documentation.xml")

	// check that file exists
	if _, err := os.Stat(kernelDocFile); os.IsNotExist(err) {
		t.Skipf("Linux Kernel documentation file not found: %s", kernelDocFile)
		return
	}

	// open file
	file, err := os.Open(kernelDocFile)
	if err != nil {
		t.Fatalf("Failed to open Linux Kernel documentation file: %v", err)
	}
	defer file.Close()

	// create parser and parse file
	parser := NewDocBookParser()
	segments, err := parser.ParseDocBook(file)

	if err != nil {
		t.Fatalf("Failed to parse Linux Kernel documentation file: %v", err)
	}

	// check that we got at least some segments
	if len(segments) == 0 {
		t.Error("Expected at least one text segment from Linux Kernel documentation file, got none")
		return
	}

	// check that segments are not empty
	for i, segment := range segments {
		if strings.TrimSpace(segment) == "" {
			t.Errorf("Segment %d is empty or contains only whitespace", i)
		}
	}

	t.Logf("Successfully parsed Linux Kernel documentation file, found %d text segments", len(segments))
}

func TestDocBookParser_FullRealDocBookFile(t *testing.T) {
	// path to DocBook in test folder
	docbookFile := filepath.Join("test", "DocBook_Definitive_Guide.xml")

	// check that file exists
	if _, err := os.Stat(docbookFile); os.IsNotExist(err) {
		t.Skipf("DocBook file not found: %s", docbookFile)
		return
	}

	// open file
	file, err := os.Open(docbookFile)
	if err != nil {
		t.Fatalf("Failed to open DocBook file: %v", err)
	}
	defer file.Close()

	// create parser and parse file
	parser := NewDocBookParser()
	segments, err := parser.ParseDocBook(file)

	if err != nil {
		t.Fatalf("Failed to parse DocBook file: %v", err)
	}

	// check that we got a significant number of segments
	if len(segments) < 100 {
		t.Errorf("Expected at least 100 text segments from DocBook file, got %d", len(segments))
		return
	}

	// check that the extracted text is of good quality
	nonEmptySegments := 0
	totalLength := 0

	for _, segment := range segments {
		trimmed := strings.TrimSpace(segment)
		if trimmed != "" {
			nonEmptySegments++
			totalLength += len(trimmed)
		}
	}

	// check that most segments are not empty
	if nonEmptySegments < len(segments)*8/10 { // 80% should be not empty
		t.Errorf("Too many empty segments: %d non-empty out of %d total", nonEmptySegments, len(segments))
	}

	// check that the average segment length is reasonable
	avgLength := totalLength / nonEmptySegments
	if avgLength < 10 {
		t.Errorf("Average segment length too short: %d characters", avgLength)
	}

	t.Logf("Successfully parsed full DocBook file:")
	t.Logf("- Total segments: %d", len(segments))
	t.Logf("- Non-empty segments: %d", nonEmptySegments)
	t.Logf("- Average segment length: %d characters", avgLength)
	t.Logf("- Total text length: %d characters", totalLength)

	// print some examples of segments
	for i, segment := range segments {
		if i >= 5 { // print only first 5 segments
			break
		}
		trimmed := strings.TrimSpace(segment)
		if trimmed != "" {
			t.Logf("Sample segment %d: %s", i, trimmed[:min(len(trimmed), 100)])
		}
	}
}
