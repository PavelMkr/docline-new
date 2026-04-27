package docline

import (
	internalAlgorithms "github.com/PavelMkr/docline-new/internal/algorithms"
	internalFramework "github.com/PavelMkr/docline-new/internal/framework"
	internalReport "github.com/PavelMkr/docline-new/internal/report"

	"fmt"
)

// Config - public configuration struct for initializing the Docline framework
type Config struct {
	ResultsDirectory    string
	DefaultReportFormat string
	DefaultTokenizer    string
	DefaultCloneFinder  string
}

type CloneFinderConfig struct {
	MinCloneLength      int
	MaxCloneLength      int
	MinGroupPower       int
	SimilarityThreshold float64
	CustomParams        map[string]interface{}
}

// FinderModeConfig type-safe public config for a API
// converts itself into the internal framework.CloneFinderConfig.
// The returned config may use CustomParams for mode-specific settings.
type FinderModeConfig interface {
	// FinderType returns the finder identifier used by the framework registry
	// (built-in: "automatic", "interactive", "heuristic", "ngram").
	FinderType() string
	toInternal(filePath string) internalFramework.CloneFinderConfig
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
func (d *Docline) AnalyzeDocument(filePath, finderType string, cfg FinderModeConfig) (*internalFramework.AnalysisResult, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil finder config")
	}

	if finderType == "" {
		finderType = cfg.FinderType()
	}
	if finderType != cfg.FinderType() {
		return nil, fmt.Errorf("finderType %q does not match config type %q", finderType, cfg.FinderType())
	}

	return d.fw.AnalyzeDocument(filePath, finderType, cfg.toInternal(filePath))
}

func (d *Docline) AnalyzeDocumentWithConfig(filePath, finderType string, cfg CloneFinderConfig) (*internalFramework.AnalysisResult, error) {
	internalCfg := internalFramework.CloneFinderConfig{
		MinCloneLength:      cfg.MinCloneLength,
		MaxCloneLength:      cfg.MaxCloneLength,
		MinGroupPower:       cfg.MinGroupPower,
		SimilarityThreshold: cfg.SimilarityThreshold,
		CustomParams:        cfg.CustomParams,
	}
	return d.fw.AnalyzeDocument(filePath, finderType, internalCfg)
}

// GenerateReport generates a report based on the analysis result
func (d *Docline) GenerateReport(result *internalFramework.AnalysisResult, format, outputPath string) error {
	return d.fw.GenerateReport(result, format, outputPath)
}
