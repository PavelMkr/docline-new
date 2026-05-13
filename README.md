# Duplicate Finder Framework
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/PavelMkr/docline-new)
![CI](https://github.com/PavelMkr/docline-new/actions/workflows/main.yml/badge.svg)

**Duplicate Finder Framework** — a clone-finding and documentation refactoring **framework** inside the DocLine project.

### How to work

- If you want the **public API**: create `docline.Docline` and call `AnalyzeDocument`, then `GenerateReport`.
- If you want the **low-level framework API**: create `framework.Framework`, register plugins manually, call `AnalyzeDocument`, then `GenerateReport`.

## Architecture

High-level scheme:

- **Public API** (`pkg/docline`):
  - `Docline`, `Config` (`docline.go`)
  - Built-in finder configs per mode (`mode_configs.go`): `AutomaticConfig`, `InteractiveConfig`, `HeuristicConfig`, `NgramConfig`
- **Framework core** (`internal/framework`):
  - `Framework`, `Config` (`core.go`)
  - `PluginRegistry` (`registry.go`)
  - Interfaces: `CloneFinder`, `DocumentParser`, `DocumentConverter`, `ReportGenerator`, `TextTokenizer`, `Filter` (`interfaces.go`)
  - Domain types: `CloneGroup`, `TextFragment`, `CloneFinderConfig`, `ReportConfig`, `AnalysisResult`, `AnalysisStatistics` (`types.go`)
- **Algorithms** (`internal/algorithms`):
  - Real implementations: automatic / interactive / openai / ngram (`*_mode.go`, `ngram_duplicate.go`)
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
- YAML (.yaml, .yml)
- HTML (.html, .htm)

*The actual "to DocBook" conversion is implemented using `pandoc` inside `internal/report.DocumentConverter`.
For simple text/DocBook files, you can work without `pandoc` (the framework will simply read the content as text or parse the DocBook directly).*

## Quickstart

An example of usage can be found in `examples/basic_usage.go`.

### Public API (`pkg/docline`)

```go
package main

import (
    "log"

    "github.com/PavelMkr/docline-new/pkg/docline"
)

func main() {
    d := docline.New(&docline.Config{
        ResultsDirectory:    "./results",
        DefaultReportFormat: "html",
        DefaultTokenizer:    "space",
        DefaultCloneFinder:  "automatic",
    })

    // Type-safe finder config for the selected mode.
    result, err := d.AnalyzeDocument("example.xml", "automatic", docline.AutomaticConfig{
        MinCloneLength: 20,
        MinGroupPower:  2,
    })
    if err != nil {
        log.Fatal(err)
    }

    if err := d.GenerateReport(result, "html", "./results/report.html"); err != nil {
        log.Fatal(err)
    }
}
```

### Low-level API (`internal/framework`)

If you need fine-grained control (custom registries/plugins), use the framework directly:

```go
package main

import (
	"log"
	"os"

	algorithms "github.com/PavelMkr/docline-new/internal/algorithms"
	"github.com/PavelMkr/docline-new/internal/framework"
	report "github.com/PavelMkr/docline-new/internal/report"
)

func main() {
	cfg := &framework.Config{
		ResultsDirectory:    "./results",
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  "automatic",
	}
	fw := framework.NewFramework(cfg)

	reg := fw.GetRegistry()

	if err := framework.RegisterBuiltInPlugins(reg); err != nil {
		log.Fatalf("register built-in plugins: %v", err)
	}
	if err := algorithms.RegisterCloneFinders(reg); err != nil {
		log.Fatalf("register clone finders: %v", err)
	}
	if err := report.RegisterDocumentPlugins(reg); err != nil {
		log.Fatalf("register document plugins: %v", err)
	}
	if err := report.RegisterReportGenerators(reg); err != nil {
		log.Fatalf("register report generators: %v", err)
	}

	docPath := "example.xml"
	if len(os.Args) > 1 {
		docPath = os.Args[1]
	}

	result, err := fw.AnalyzeDocument(docPath, "automatic", framework.CloneFinderConfig{
		MinCloneLength: 20,
		MinGroupPower:  2,
	})
	if err != nil {
		log.Fatalf("analyze document: %v", err)
	}

	if err := fw.GenerateReport(result, "html", "./results/report.html"); err != nil {
		log.Fatalf("generate report: %v", err)
	}

	log.Println("Done: ./results/report.html")
}
```

And add into go.mod as:

`replace github.com/PavelMkr/docline-new v0.0.0-unpublished => ../docline`

## Framework extension

- **Your own clone finder algorithm**: implement the `CloneFinder` interface and register it via `PluginRegistry.RegisterCloneFinder`.
  - Example: `examples/custom_finder.go` and `examples/custom_finder/main.go`.
- **Your own report generator**: implement the `ReportGenerator` interface and register it via `RegisterReportGenerator`.
  - Example: `examples/custom_report.go` and `examples/custom_report/main.go`.

## Dependencies

- Go **1.23+**
- Optional: **Pandoc** is only needed for converting input documents to DocBook (via `DocumentConverter`).
  The tests in `tests/converter_test.go` and the functionality of `PandocConverterAdapter` assume its presence, but basic analysis of DocBook/XML and plain text works without it.
