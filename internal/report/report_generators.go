package internal

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PavelMkr/docline-new/internal/framework"
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

	totalTokens := totalTokensFromSettings(cfg.Settings)
	if totalTokens <= 0 {
		totalTokens = maxEndPos(groups)
	}
	heat := buildHeatmap(groups, totalTokens, 120)

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\">")
	sb.WriteString("<title>" + htmlEscape(cfg.Title) + "</title>")
	sb.WriteString("<style>body{font-family:sans-serif}table{border-collapse:collapse;width:100%}th,td{border:1px solid #ccc;padding:4px;font-size:14px}code{white-space:pre-wrap}.heatmap{display:flex;gap:1px;height:14px;align-items:stretch}.heat{flex:1 1 auto}</style>")
	sb.WriteString("</head><body>")
	sb.WriteString("<h1>" + htmlEscape(cfg.Title) + "</h1>")
	if cfg.SourceFile != "" {
		sb.WriteString("<p><b>Source:</b> " + htmlEscape(cfg.SourceFile) + "</p>")
	}
	sb.WriteString(fmt.Sprintf("<p><b>Total groups:</b> %d</p>", len(groups)))
	if len(heat) > 0 && totalTokens > 0 {
		sb.WriteString("<p><b>Heatmap:</b> (token coverage, bins=" + fmt.Sprint(len(heat)) + ", total_tokens=" + fmt.Sprint(totalTokens) + ")</p>")
		sb.WriteString(renderHeatmapHTML(heat))
	}

	sb.WriteString("<table><thead><tr><th>#</th><th>Power</th><th>Archetype</th><th>Fragments</th></tr></thead><tbody>")
	for i, g := range groups {
		sb.WriteString("<tr>")
		sb.WriteString(fmt.Sprintf("<td>%d</td>", i+1))
		sb.WriteString(fmt.Sprintf("<td>%d</td>", g.Power))
		sb.WriteString("<td><code>" + htmlEscape(g.Archetype) + "</code></td>")
		sb.WriteString("<td><ol>")
		for _, f := range g.Fragments {
			prefix := ""
			if ln1, ok1 := intFromMetadata(f.Metadata, "source_line_start"); ok1 && ln1 > 0 {
				ln2, ok2 := intFromMetadata(f.Metadata, "source_line_end")
				if ok2 && ln2 > 0 && ln2 != ln1 {
					prefix = "L" + fmt.Sprint(ln1) + "-" + fmt.Sprint(ln2) + ": "
				} else {
					prefix = "L" + fmt.Sprint(ln1) + ": "
				}
			}
			sb.WriteString("<li><code>" + htmlEscape(prefix+f.Content) + "</code></li>")
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

	totalTokens := totalTokensFromSettings(cfg.Settings)
	if totalTokens <= 0 {
		totalTokens = maxEndPos(groups)
	}
	heat := buildHeatmap(groups, totalTokens, 120)
	stats := statsFromSettings(cfg.Settings)

	payload := struct {
		Title      string                       `json:"title"`
		SourceFile string                       `json:"source_file"`
		Settings   map[string]interface{}       `json:"settings,omitempty"`
		Groups     []framework.CloneGroup       `json:"groups"`
		Stats      framework.AnalysisStatistics `json:"stats,omitempty"`
		Heatmap    []int                        `json:"heatmap,omitempty"`
		TotalTokens int                         `json:"total_tokens,omitempty"`
	}{
		Title:      cfg.Title,
		SourceFile: cfg.SourceFile,
		Settings:   cfg.Settings,
		Groups:     groups,
		Stats:      stats,
		Heatmap:    heat,
		TotalTokens: totalTokens,
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

func totalTokensFromSettings(settings map[string]interface{}) int {
	if settings == nil {
		return 0
	}
	if v, ok := settings["total_tokens"]; ok {
		switch vv := v.(type) {
		case int:
			return vv
		case int32:
			return int(vv)
		case int64:
			return int(vv)
		case float32:
			return int(vv)
		case float64:
			return int(vv)
		}
	}
	return 0
}

func statsFromSettings(settings map[string]interface{}) framework.AnalysisStatistics {
	if settings == nil {
		return framework.AnalysisStatistics{}
	}
	v, ok := settings["stats"]
	if !ok || v == nil {
		return framework.AnalysisStatistics{}
	}
	switch vv := v.(type) {
	case framework.AnalysisStatistics:
		return vv
	case *framework.AnalysisStatistics:
		if vv == nil {
			return framework.AnalysisStatistics{}
		}
		return *vv
	default:
		// Unknown representation (e.g. after JSON roundtrip). Keep zero value.
		return framework.AnalysisStatistics{}
	}
}

func maxEndPos(groups []framework.CloneGroup) int {
	maxEnd := 0
	for _, g := range groups {
		for _, f := range g.Fragments {
			if f.EndPos > maxEnd {
				maxEnd = f.EndPos
			}
		}
	}
	return maxEnd
}

// buildHeatmap creates a fixed-width histogram over token positions.
// heat[i] == amount of token coverage that falls into bin i.
func buildHeatmap(groups []framework.CloneGroup, totalTokens int, bins int) []int {
	if totalTokens <= 0 || bins <= 0 {
		return nil
	}
	// If bins > totalTokens, integer division would create many zero-width bins,
	// making overlaps impossible to count. Clamp bins to ensure each bin spans
	// at least one token.
	if bins > totalTokens {
		bins = totalTokens
	}
	heat := make([]int, bins)
	for _, g := range groups {
		for _, f := range g.Fragments {
			start := f.StartPos
			end := f.EndPos
			if start < 0 {
				start = 0
			}
			if end > totalTokens {
				end = totalTokens
			}
			if end <= start {
				continue
			}

			// Add overlap length to bins without per-token loop.
			// Bin i covers [i*totalTokens/bins, (i+1)*totalTokens/bins).
			startBin := (start * bins) / totalTokens
			endBin := ((end - 1) * bins) / totalTokens
			if startBin < 0 {
				startBin = 0
			}
			if endBin >= bins {
				endBin = bins - 1
			}
			for bi := startBin; bi <= endBin; bi++ {
				bs := (bi * totalTokens) / bins
				be := ((bi + 1) * totalTokens) / bins
				ov := overlapLen(start, end, bs, be)
				if ov > 0 {
					heat[bi] += ov
				}
			}
		}
	}
	return heat
}

func overlapLen(aStart, aEnd, bStart, bEnd int) int {
	if aEnd <= bStart || bEnd <= aStart {
		return 0
	}
	s := aStart
	if bStart > s {
		s = bStart
	}
	e := aEnd
	if bEnd < e {
		e = bEnd
	}
	return e - s
}

func renderHeatmapHTML(heat []int) string {
	maxV := 0
	for _, v := range heat {
		if v > maxV {
			maxV = v
		}
	}
	if maxV == 0 {
		return "<div class=\"heatmap\"></div>"
	}

	var sb strings.Builder
	sb.WriteString("<div class=\"heatmap\" aria-label=\"heatmap\">")
	for _, v := range heat {
		intensity := float64(v) / float64(maxV) // 0..1
		// Blue-ish gradient; keep background visible even at 0.
		alpha := 0.05 + 0.95*intensity
		sb.WriteString("<div class=\"heat\" style=\"background-color: rgba(30, 136, 229, " + fmt.Sprintf("%.3f", alpha) + ")\"></div>")
	}
	sb.WriteString("</div>")
	return sb.String()
}

func intFromMetadata(m map[string]interface{}, key string) (int, bool) {
	if m == nil {
		return 0, false
	}
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch vv := v.(type) {
	case int:
		return vv, true
	case int32:
		return int(vv), true
	case int64:
		return int(vv), true
	case float32:
		return int(vv), true
	case float64:
		return int(vv), true
	default:
		return 0, false
	}
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
