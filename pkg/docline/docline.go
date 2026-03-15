package docline

import (
    internalFramework "github.com/PavelMkr/docline-new/internal/framework"
    internalAlgorithms "github.com/PavelMkr/docline-new/internal/algorithms"
    internalReport "github.com/PavelMkr/docline-new/internal/report"
)

// Config - public configuration struct for initializing the Docline framework
type Config struct {
    ResultsDirectory    string
    DefaultReportFormat string
    DefaultTokenizer    string
    DefaultCloneFinder  string
}

// Docline - main struct for the Docline framework
type Docline struct {
    fw *internalFramework.Framework
}

// New creates a new Docline instance
func New(cfg *Config) *Docline {
    internalCfg := &internalFramework.Config{
        ResultsDirectory:    cfg.ResultsDirectory,
        DefaultReportFormat: cfg.DefaultReportFormat,
        DefaultTokenizer:    cfg.DefaultTokenizer,
        DefaultCloneFinder:  cfg.DefaultCloneFinder,
    }

    fw := internalFramework.NewFramework(internalCfg)

    // Регистрируем плагины
    reg := fw.GetRegistry()
    internalFramework.RegisterBuiltInPlugins(reg)
    internalAlgorithms.RegisterCloneFinders(reg)
    internalReport.RegisterDocumentPlugins(reg)
    internalReport.RegisterReportGenerators(reg)

    return &Docline{fw: fw}
}

// AnalyzeDocument analyzes the specified document and returns the result
func (d *Docline) AnalyzeDocument(filePath, finderType string, minCloneLength, minGroupPower int) (*internalFramework.AnalysisResult, error) {
    cfg := internalFramework.CloneFinderConfig{
        MinCloneLength: minCloneLength,
        MinGroupPower:  minGroupPower,
    }
    return d.fw.AnalyzeDocument(filePath, finderType, cfg)
}

// GenerateReport generates a report based on the analysis result
func (d *Docline) GenerateReport(result *internalFramework.AnalysisResult, format, outputPath string) error {
    return d.fw.GenerateReport(result, format, outputPath)
}
