# Duplicate Finder Framework
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/PavelMkr/docline-new)
![CI](https://github.com/PavelMkr/docline-new/actions/workflows/main.yml/badge.svg)

**Duplicate Finder Framework** — a clone-finding and documentation refactoring **framework** inside the DocLine project.

### How to work

- Create an instance of `framework.Framework`.
- Register the required plugins (algorithms, document parser/converter, report generators, tokenizers, and filters).
- Call `AnalyzeDocument` (or `AnalyzeDocumentWithConfig` for `heuristic`)  and then `GenerateReport` in the desired format.

## Architecture

High-level scheme:

- **Framework core** (`internal/framework`):
  - `Framework`, `Config` (`core.go`)
  - `PluginRegistry` (`registry.go`)
  - Interfaces: `CloneFinder`, `DocumentParser`, `DocumentConverter`, `ReportGenerator`, `TextTokenizer`, `Filter` (`interfaces.go`)
  - Domain types: `CloneGroup`, `TextFragment`, `CloneFinderConfig`, `ReportConfig`, `AnalysisResult`, `AnalysisStatistics` (`types.go`)
- **Algorithms** (`internal/algorithms`):
  - Real implementations: automatic / interactive / heuristic / ngram (`*_mode.go`, `ngram_duplicate.go`)
  - Adapters for `CloneFinder`: `AutomaticModeAdapter`, `InteractiveModeAdapter`, `NGramAdapter` (`framework_adapters.go`)
- **Document parser and converter** (`internal/report`):
  - `DocBookParser`, `NewDocBookParser` (`docbook_parser.go`)
  - `DocumentConverter`, `NewDocumentConverter` (`converter.go`)
  - Adapters for `DocumentParser`/`DocumentConverter`: `DocBookParserAdapter`, `PandocConverterAdapter` (`framework_adapters.go`)
- **Report generators** (`internal/report/report_generators.go`):
  - Plugin implementations of `HTMLReportGenerator`, `JSONReportGenerator`, `CSVReportGenerator`.
- **Utilities / core plugins** (`internal/framework/adapters.go`, `builtins.go`):
  - `SpaceTokenizer`, `StrictFilter`, `JaccardSimilarityCalculator`
  - Registration via `framework.RegisterBuiltInPlugins(registry)`.

## Supported file formats

- DocBook (.xml, .dbk, .docbook)
- Microsoft Word (.doc, .docx)
- OpenDocument Text (.odt)
- Rich Text Format (.rtf)
- Markdown (.md)
- Plain Text (.txt)
- HTML (.html, .htm)

*The actual "to DocBook" conversion is implemented using `pandoc` inside `internal/report.DocumentConverter`.
For simple text/DocBook files, you can work without `pandoc` (the framework will simply read the content as text or parse the DocBook directly).*

## Quickstart

An example of usage can be found in `examples/basic_usage.go`. The schema is as follows:

```go
cfg := &framework.Config{
    ResultsDirectory:    "./results",
    DefaultReportFormat: "html",
    DefaultTokenizer:    "space",
    DefaultCloneFinder:  "automatic",
}
fw := framework.NewFramework(cfg)

reg := fw.GetRegistry()

// 1. Register the basic framework utilities
_ = framework.RegisterBuiltInPlugins(reg)

// 2. Register built-in algorithms, a parser/converter, and report generators
_ = algorithms.RegisterCloneFinders(reg)
_ = report.RegisterDocumentPlugins(reg)
_ = report.RegisterReportGenerators(reg)

// 3. Start document analysis
result, err := fw.AnalyzeDocument("example.xml", "automatic", framework.CloneFinderConfig{
    MinCloneLength: 20,
    MinGroupPower:  2,
})

// 4. Generate a report
err = fw.GenerateReport(result, "html", "./results/report.html")
```

## Framework extension

- **Your own clone finder algorithm**: implement the `CloneFinder` interface and register it via `PluginRegistry.RegisterCloneFinder`.
  - Example: `examples/custom_finder.go` and `examples/custom_finder/main.go`.
- **Your own report generator**: implement the `ReportGenerator` interface and register it via `RegisterReportGenerator`.
  - Example: `examples/custom_report.go` and `examples/custom_report/main.go`.

## Dependencies

- Go **1.23+**
- Optional: **Pandoc** is only needed for converting input documents to DocBook (via `DocumentConverter`).
  The tests in `tests/converter_test.go` and the functionality of `PandocConverterAdapter` assume its presence, but basic analysis of DocBook/XML and plain text works without it.
