package main

import "testing"

func TestIsConversionNeeded(t *testing.T) {
	conv := NewDocumentConverter()
	if !conv.IsConversionNeeded("file.docx") {
		t.Error("expected conversion needed for .docx")
	}
	if conv.IsConversionNeeded("file.xml") {
		t.Error("expected no conversion needed for .xml")
	}
}

func TestGetPandocFormat(t *testing.T) {
	cases := []struct {
		ext    string
		expect string
	}{
		{".doc", "docx"},
		{".docx", "docx"},
		{".odt", "odt"},
		{".rtf", "rtf"},
		{".md", "markdown"},
		{".txt", "plain"},
		{".html", "html"},
		{".htm", "html"},
		{".unknown", "plain"},
	}
	for _, c := range cases {
		if got := getPandocFormat(c.ext); got != c.expect {
			t.Errorf("for ext %s expected %s, got %s", c.ext, c.expect, got)
		}
	}
}
