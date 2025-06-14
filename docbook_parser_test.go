package main

import (
	"encoding/xml"
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
