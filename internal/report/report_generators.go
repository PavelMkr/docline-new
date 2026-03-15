package internal

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Docline/internal/framework"
)

// HTMLReportGenerator implements framework.ReportGenerator for HTML output.
type HTMLReportGenerator struct{}

func (h *HTMLReportGenerator) Name() string {
	return "html-report"
}

func (h *HTMLReportGenerator) Format() string {
	return "html"
}

func (h *HTMLReportGenerator) Generate(groups []framework.CloneGroup, cfg framework.ReportConfig, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\">")
	sb.WriteString("<title>" + htmlEscape(cfg.Title) + "</title>")
	sb.WriteString("<style>body{font-family:sans-serif}table{border-collapse:collapse;width:100%}th,td{border:1px solid #ccc;padding:4px;font-size:14px}code{white-space:pre-wrap}</style>")
	sb.WriteString("</head><body>")
	sb.WriteString("<h1>" + htmlEscape(cfg.Title) + "</h1>")
	if cfg.SourceFile != "" {
		sb.WriteString("<p><b>Source:</b> " + htmlEscape(cfg.SourceFile) + "</p>")
	}
	sb.WriteString(fmt.Sprintf("<p><b>Total groups:</b> %d</p>", len(groups)))

	sb.WriteString("<table><thead><tr><th>#</th><th>Power</th><th>Archetype</th><th>Fragments</th></tr></thead><tbody>")
	for i, g := range groups {
		sb.WriteString("<tr>")
		sb.WriteString(fmt.Sprintf("<td>%d</td>", i+1))
		sb.WriteString(fmt.Sprintf("<td>%d</td>", g.Power))
		sb.WriteString("<td><code>" + htmlEscape(g.Archetype) + "</code></td>")
		sb.WriteString("<td><ol>")
		for _, f := range g.Fragments {
			sb.WriteString("<li><code>" + htmlEscape(f.Content) + "</code></li>")
		}
		sb.WriteString("</ol></td>")
		sb.WriteString("</tr>")
	}
	sb.WriteString("</tbody></table></body></html>")

	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

// JSONReportGenerator implements framework.ReportGenerator for JSON output.
type JSONReportGenerator struct{}

func (j *JSONReportGenerator) Name() string {
	return "json-report"
}

func (j *JSONReportGenerator) Format() string {
	return "json"
}

func (j *JSONReportGenerator) Generate(groups []framework.CloneGroup, cfg framework.ReportConfig, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	payload := struct {
		Title      string                       `json:"title"`
		SourceFile string                       `json:"source_file"`
		Settings   map[string]interface{}       `json:"settings,omitempty"`
		Groups     []framework.CloneGroup       `json:"groups"`
		Stats      framework.AnalysisStatistics `json:"stats,omitempty"`
	}{
		Title:      cfg.Title,
		SourceFile: cfg.SourceFile,
		Settings:   cfg.Settings,
		Groups:     groups,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	return os.WriteFile(outputPath, data, 0o644)
}

// CSVReportGenerator implements framework.ReportGenerator for CSV output
// similar in spirit to the old WriteShortTermsCSV helper.
type CSVReportGenerator struct {
	// MaxTokens limits fragment length to be included; 0 means no limit.
	MaxTokens int
	// MinOccurs specifies minimal number of fragments per group.
	MinOccurs int
}

func (c *CSVReportGenerator) Name() string {
	return "csv-report"
}

func (c *CSVReportGenerator) Format() string {
	return "csv"
}

func (c *CSVReportGenerator) Generate(groups []framework.CloneGroup, cfg framework.ReportConfig, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create csv file: %w", err)
	}
	defer file.Close()

	// Use UTF-8 BOM so that Excel opens the file correctly.
	if _, err := file.Write([]byte("\uFEFF")); err != nil {
		return fmt.Errorf("write bom: %w", err)
	}

	w := csv.NewWriter(file)
	w.Comma = ';'

	if err := w.Write([]string{"N tokens", "Occurs times", "Text"}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	maxTokens := c.MaxTokens
	if maxTokens == 0 {
		maxTokens = 10
	}
	minOccurs := c.MinOccurs
	if minOccurs == 0 {
		minOccurs = 2
	}

	for _, g := range groups {
		if len(g.Fragments) == 0 {
			continue
		}
		ntoks := len(strings.Fields(g.Fragments[0].Content))
		if ntoks <= maxTokens && len(g.Fragments) >= minOccurs {
			txt := strings.ReplaceAll(g.Fragments[0].Content, "\n", " ")
			txt = strings.ReplaceAll(txt, ";", ",")
			record := []string{fmt.Sprint(ntoks), fmt.Sprint(len(g.Fragments)), txt}
			if err := w.Write(record); err != nil {
				return fmt.Errorf("write record: %w", err)
			}
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}
	return nil
}

// RegisterReportGenerators registers built-in HTML/JSON/CSV report generators.
func RegisterReportGenerators(reg *framework.PluginRegistry) error {
	if err := reg.RegisterReportGenerator(&HTMLReportGenerator{}); err != nil {
		return fmt.Errorf("register html report generator: %w", err)
	}
	if err := reg.RegisterReportGenerator(&JSONReportGenerator{}); err != nil {
		return fmt.Errorf("register json report generator: %w", err)
	}
	if err := reg.RegisterReportGenerator(&CSVReportGenerator{MaxTokens: 3, MinOccurs: 2}); err != nil {
		return fmt.Errorf("register csv report generator: %w", err)
	}
	return nil
}
