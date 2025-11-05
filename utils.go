package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GenerateNGrams creates n-grams from input text.
func GenerateNGrams(text string, n int) []string {
	words := strings.Fields(text)
	var ngrams []string
	for i := 0; i <= len(words)-n; i++ {
		ngrams = append(ngrams, strings.Join(words[i:i+n], " "))
	}
	return ngrams
}

// writeToFile writes data to a file at the specified path
func writeToFile(filePath string, data string) error {
	fmt.Printf("writeToFile: Attempting to write to %s\n", filePath)
	fmt.Printf("writeToFile: Data length: %d bytes\n", len(data))

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	fmt.Printf("writeToFile: Ensuring directory exists: %s\n", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("writeToFile: Failed to create directory: %v\n", err)
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create or truncate the file
	fmt.Printf("writeToFile: Creating/truncating file\n")
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("writeToFile: Failed to create file: %v\n", err)
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write data to file
	fmt.Printf("writeToFile: Writing data to file\n")
	bytesWritten, err := file.WriteString(data)
	if err != nil {
		fmt.Printf("writeToFile: Failed to write data: %v\n", err)
		return fmt.Errorf("failed to write data: %v", err)
	}
	fmt.Printf("writeToFile: Successfully wrote %d bytes to file\n", bytesWritten)

	// Ensure data is written to disk
	if err := file.Sync(); err != nil {
		fmt.Printf("writeToFile: Failed to sync file: %v\n", err)
		return fmt.Errorf("failed to sync file: %v", err)
	}

	fmt.Printf("writeToFile: File write completed successfully\n")
	return nil
}

// writeJSON writes any value as pretty JSON to file
func writeJSON(filePath string, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %v", err)
	}
	return writeToFile(filePath, string(b))
}

// writeSimpleHTML writes a minimal HTML page with provided title and body HTML
func writeSimpleHTML(filePath, title, bodyHTML string) error {
	html := "" +
		"<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"utf-8\"><title>" +
		title +
		"</title><style>table{border-collapse:collapse}td,th{border:1px solid #ccc;padding:4px;font-family:sans-serif}</style></head><body>" +
		"<h2>" + title + "</h2>" + bodyHTML + "</body></html>"
	return writeToFile(filePath, html)
}

// htmlEscape performs minimal HTML escaping for text content
func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

// writeTextAsHTML wraps plain text in a <pre> and writes as HTML
func writeTextAsHTML(filePath, title, text string) error {
	body := "<pre style=\"white-space:pre-wrap;word-wrap:break-word;font-family:monospace\">" + htmlEscape(text) + "</pre>"
	return writeSimpleHTML(filePath, title, body)
}

// WritePygroupsHTML renders a simple groups table similar to clones2html.py output
func WritePygroupsHTML(targetPath string, groups []CloneGroup, filenames []string, avgTok float64, dirtyGrp int) error {
	// files list
	var sb strings.Builder
	if len(filenames) > 0 {
		sb.WriteString("<p>Files:<br/>")
		for i, f := range filenames {
			if i > 0 {
				sb.WriteString("<br/>")
			}
			sb.WriteString(strings.ReplaceAll(f, "&", "&amp;"))
		}
		sb.WriteString("</p>")
	}
	sb.WriteString(fmt.Sprintf("<p>Avg tokens in clone: %.3f</p>", avgTok))
	sb.WriteString(fmt.Sprintf("<p>Dirty groups: %d</p>", dirtyGrp))
	sb.WriteString("<table><thead><tr><th>#</th><th>#Tokens</th><th>Occurs</th><th>Text</th></tr></thead><tbody>")
	for i, g := range groups {
		ntoks := 0
		if len(g.Fragments) > 0 {
			ntoks = len(strings.Fields(g.Fragments[0].Content))
		}
		occurs := len(g.Fragments)
		// take first fragment content as representative
		text := ""
		if len(g.Fragments) > 0 {
			text = g.Fragments[0].Content
		}
		esc := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\n", "<br/>")
		sb.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%d</td><td>%d</td><td><tt>%s</tt></td></tr>", i+1, ntoks, occurs, esc.Replace(text)))
	}
	sb.WriteString("</tbody></table>")
	return writeSimpleHTML(targetPath, "Clone groups", sb.String())
}

// WritePyVariativeElements renders a minimal interactive-like report placeholder
func WritePyVariativeElements(targetPath string, groups []CloneGroup) error {
	var sb strings.Builder
	sb.WriteString("<p>Variative elements (simplified)</p>")
	for i, g := range groups {
		sb.WriteString(fmt.Sprintf("<h3>Group %d (power %d)</h3>", i+1, g.Power))
		sb.WriteString("<ul>")
		for _, f := range g.Fragments {
			esc := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
			sb.WriteString("<li><code>" + esc.Replace(f.Content) + "</code></li>")
		}
		sb.WriteString("</ul>")
	}
	// include client-side placeholder js for future interactivity
	sb.WriteString("<script>console.log('pyvarelements simplified report loaded');</script>")
	return writeSimpleHTML(targetPath, "Variative Elements", sb.String())
}

// WriteDensityReports writes placeholder density visualizations
func WriteDensityReports(dir string, totalTokens int, groups []CloneGroup) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if totalTokens < 0 {
		totalTokens = 0
	}
	density := make([]int, totalTokens)
	maxv := 0
	for _, g := range groups {
		for _, f := range g.Fragments {
			b := f.StartPos
			e := f.EndPos
			if b < 0 {
				b = 0
			}
			if e > totalTokens {
				e = totalTokens
			}
			for i := b; i < e; i++ {
				density[i]++
				if density[i] > maxv {
					maxv = density[i]
				}
			}
		}
	}
	if maxv == 0 {
		maxv = 1
	}

	bins := 600
	if totalTokens < bins {
		bins = totalTokens
	}
	if bins == 0 {
		bins = 1
	}
	binSize := (totalTokens + bins - 1) / bins

	var tableSB strings.Builder
	tableSB.WriteString("<table><thead><tr><th>#</th><th>Start</th><th>End</th><th>Avg density</th></tr></thead><tbody>")
	var mapSB strings.Builder
	mapSB.WriteString("<div style=\"white-space:nowrap;border:1px solid #ccc;height:24px\">")

	for i, bi := 0, 0; i < totalTokens; i += binSize {
		end := i + binSize
		if end > totalTokens {
			end = totalTokens
		}
		sum := 0
		for j := i; j < end; j++ {
			sum += density[j]
		}
		avg := float64(sum)
		if end-i > 0 {
			avg = avg / float64(end-i)
		} else {
			avg = 0
		}
		alpha := avg / float64(maxv)
		if alpha > 1 {
			alpha = 1
		}
		widthPct := float64(end-i) / float64(maxInt(totalTokens, 1)) * 100.0
		tableSB.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%d</td><td>%d</td><td>%.3f</td></tr>", bi+1, i, end, avg))
		mapSB.WriteString(fmt.Sprintf("<span title=\"tokens %d-%d, avg=%.2f\" style=\"display:inline-block;height:24px;width:%.4f%%;background:rgba(0,128,255,%.3f)\"></span>", i, end, avg, widthPct, alpha))
		bi++
	}
	tableSB.WriteString("</tbody></table>")
	mapSB.WriteString("</div>")

	if err := writeSimpleHTML(filepath.Join(dir, "densityreport.html"), "Density report", tableSB.String()); err != nil {
		return err
	}
	if err := writeSimpleHTML(filepath.Join(dir, "densitymap.html"), "Density map", mapSB.String()); err != nil {
		return err
	}

	var heatSB strings.Builder
	heatSB.WriteString("<div style=\"white-space:nowrap;border:1px solid #ccc;height:24px\">")
	for i := 0; i < totalTokens; i += binSize {
		end := i + binSize
		if end > totalTokens {
			end = totalTokens
		}
		sum := 0
		for j := i; j < end; j++ {
			sum += density[j]
		}
		avg := float64(sum)
		if end-i > 0 {
			avg = avg / float64(end-i)
		} else {
			avg = 0
		}
		alpha := avg / float64(maxv)
		if alpha > 1 {
			alpha = 1
		}
		widthPct := float64(end-i) / float64(maxInt(totalTokens, 1)) * 100.0
		heatSB.WriteString(fmt.Sprintf("<span title=\"tokens %d-%d, avg=%.2f\" style=\"display:inline-block;height:24px;width:%.4f%%;background:rgba(255,0,0,%.3f)\"></span>", i, end, avg, widthPct, alpha))
	}
	heatSB.WriteString("</div>")
	if err := writeSimpleHTML(filepath.Join(dir, "heatmap.html"), "Heatmap", heatSB.String()); err != nil {
		return err
	}
	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// WriteShortTermsCSV writes a basic short terms CSV for groups with <= maxTokens
func WriteShortTermsCSV(targetPath string, groups []CloneGroup, maxTokens, minOccurs int) error {
	var sb strings.Builder
	sb.WriteString("\uFEFF") // BOM for Excel friendliness
	sb.WriteString("N tokens;Occurs times;Text\n")
	for _, g := range groups {
		if len(g.Fragments) == 0 {
			continue
		}
		ntoks := len(strings.Fields(g.Fragments[0].Content))
		if ntoks <= maxTokens && len(g.Fragments) >= minOccurs {
			// collapse text to one line
			txt := strings.ReplaceAll(g.Fragments[0].Content, "\n", " ")
			txt = strings.ReplaceAll(txt, ";", ",")
			sb.WriteString(fmt.Sprintf("%d;%d;%s\n", ntoks, len(g.Fragments), txt))
		}
	}
	return writeToFile(targetPath, sb.String())
}

// AverageTokensInGroup computes average token length over all fragments
func AverageTokensInGroup(groups []CloneGroup) float64 {
	total := 0
	count := 0
	for _, g := range groups {
		for _, f := range g.Fragments {
			total += len(strings.Fields(f.Content))
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}

// readFileContent reads file content and handles different file types
func readFileContent(filePath string) (string, error) {
	fmt.Printf("Reading file content: %s\n", filePath)
	converter := NewDocumentConverter()
	var tempFilePath string
	var err error

	// Check if conversion is needed
	if converter.IsConversionNeeded(filePath) {
		fmt.Printf("Conversion needed for file: %s\n", filePath)
		tempFilePath, err = converter.ConvertToDocBook(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to convert file: %v", err)
		}
		defer converter.CleanupTempFile(tempFilePath)
		filePath = tempFilePath
	} else {
		fmt.Printf("No conversion needed for file: %s\n", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	fmt.Printf("Successfully read file content, size: %d bytes\n", len(content))

	// Check if file is DocBook/XML
	ext := strings.ToLower(filepath.Ext(filePath))
	fmt.Printf("File extension: %s\n", ext)
	if ext == ".xml" || ext == ".dbk" || ext == ".docbook" {
		fmt.Printf("Processing as DocBook/XML file\n")
		parser := NewDocBookParser()
		segments, err := parser.ParseDocBook(strings.NewReader(string(content)))
		if err != nil {
			return "", fmt.Errorf("failed to parse DocBook: %v", err)
		}
		fmt.Printf("Successfully parsed DocBook, found %d segments\n", len(segments))
		// Join segments with newlines
		return strings.Join(segments, "\n"), nil
	}

	fmt.Printf("Processing as regular text file\n")
	// For other file types, return content as is
	return string(content), nil
}

// splitTextIntoParts splits text into parts (for example, by sentences).
func splitTextIntoParts(text string) []string {
	// Use regex to find punctuation marks.
	re := regexp.MustCompile(`[.!?]`)
	parts := re.Split(text, -1)

	// Remove empty lines and whitespace.
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// validateFileFormat checks if the file format is supported
func validateFileFormat(filePath string) error {
	converter := NewDocumentConverter()
	ext := strings.ToLower(filepath.Ext(filePath))

	fmt.Printf("Validating file format: %s (extension: %s)\n", filePath, ext)
	fmt.Printf("Is supported output format: %v\n", converter.SupportedOutputFormats[ext])
	fmt.Printf("Is supported input format: %v\n", converter.SupportedInputFormats[ext])

	// check, is the file DocBook or a supported format for conversion
	if !converter.SupportedOutputFormats[ext] && !converter.SupportedInputFormats[ext] {
		return fmt.Errorf("unsupported file format: %s. Supported formats are: DocBook (.xml, .dbk, .docbook), Microsoft Word (.doc, .docx), OpenDocument (.odt), RTF (.rtf), Markdown (.md), Plain Text (.txt), HTML (.html, .htm)", ext)
	}
	return nil
}
