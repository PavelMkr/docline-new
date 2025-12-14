package main

import (
	"fmt"
	"os"
	"strings"

	"Docline/framework"
)

// MarkdownReportGenerator implements a custom report generator for Markdown format
type MarkdownReportGenerator struct{}

func (m *MarkdownReportGenerator) Name() string {
	return "markdown"
}

func (m *MarkdownReportGenerator) Format() string {
	return "md"
}

func (m *MarkdownReportGenerator) Generate(groups []framework.CloneGroup, config framework.ReportConfig, outputPath string) error {
	var sb strings.Builder
	
	// Write header
	sb.WriteString("# " + config.Title + "\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** %s\n\n", config.SourceFile))
	
	// Write groups
	sb.WriteString(fmt.Sprintf("## Found %d Clone Groups\n\n", len(groups)))
	
	for i, group := range groups {
		sb.WriteString(fmt.Sprintf("### Group %d (Power: %d)\n\n", i+1, group.Power))
		sb.WriteString(fmt.Sprintf("**Archetype:** `%s`\n\n", group.Archetype))
		sb.WriteString("**Fragments:**\n\n")
		
		for j, frag := range group.Fragments {
			sb.WriteString(fmt.Sprintf("%d. [%d-%d] `%s`\n", j+1, frag.StartPos, frag.EndPos, frag.Content))
		}
		
		sb.WriteString("\n---\n\n")
	}
	
	// Write to file
	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

func main() {
	// Create framework
	fw := framework.NewFramework(nil)
	
	// Register custom report generator
	mdGen := &MarkdownReportGenerator{}
	err := fw.GetRegistry().RegisterReportGenerator(mdGen)
	if err != nil {
		panic(err)
	}
	
	// Create sample result
	result := &framework.AnalysisResult{
		Groups: []framework.CloneGroup{
			{
				Fragments: []framework.TextFragment{
					{Content: "This is a duplicate", StartPos: 0, EndPos: 4},
					{Content: "This is a duplicate", StartPos: 100, EndPos: 104},
				},
				Power:     2,
				Archetype: "This is a duplicate",
			},
		},
		Statistics: framework.AnalysisStatistics{
			TotalGroups:    1,
			TotalFragments: 2,
		},
	}
	
	// Generate markdown report
	err = fw.GenerateReport(result, "md", "./results/report.md")
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Markdown report generated!")
}

