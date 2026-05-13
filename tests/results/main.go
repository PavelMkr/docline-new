package main

import (
	"log"
	"os"

	"github.com/PavelMkr/docline-new/pkg/docline"
)

func automatic(d *docline.Docline) {
  convertToDRL := true
  archetypeLen := 6
  strict := true

  result, err := d.AnalyzeDocument("documentations/openapiOllama.yaml", "automatic", docline.AutomaticConfig{
    MinCloneLength:  20,
    MinGroupPower:   2,
    ConvertToDRL:    &convertToDRL,
    ArchetypeLength: &archetypeLen,
    StrictFilter:    &strict,
  })
  if err != nil {
    log.Fatal(err)
  }
  if err := d.GenerateReport(result, "html", "./results/automatic1.html"); err != nil {
    log.Fatal(err)
  }
}

func heuristic(d *docline.Docline) {
  result, err := d.AnalyzeDocument("documentations/DocBook_Definitive_Guide.xml", "heuristic", docline.HeuristicConfig{
    MinCloneLength:  3,
    MinGroupPower:   2,
  })
  if err != nil {
    log.Fatal(err)
  }
  if err := d.GenerateReport(result, "csv", "./results/heuristic.csv"); err != nil {
    log.Fatal(err)
  }
}



func ngram(d *docline.Docline) {
  result, err := d.AnalyzeDocument("documentations/openapi-with-code-samples-OpenAI.yml", "ngram", docline.NgramConfig{
    MinCloneLength: 20,
    MinGroupPower:  2,
    MaxEdit:        3,
    MaxFuzzy:       2, 
  })
  if err != nil {
    log.Fatal(err)
  }
  if err := d.GenerateReport(result, "html", "./results/ngram.html"); err != nil {
    log.Fatal(err)
  }
}

func openai(d *docline.Docline) {
   result, err := d.AnalyzeDocument("documentations/Linux_Kernel_Documentation.xml", "openai", docline.OpenAIConfig{
    MinGroupPower: 2,
    Provider:      "openai",
    Model:         "gpt-4o-mini",
    Prompt:        "Найди нечёткие повторы и сгруппируй по смыслу",
    APIKey:        os.Getenv("OPENAI_API_KEY"),
    BaseURL:       "https://api.openai.com/v1", // Для ollama: http://localhost:11434/v1
    TimeoutSec:    90,
  })
  if err != nil {
    log.Fatal(err)
  }

  if err := d.GenerateReport(result, "html", "./results/openai.html"); err != nil {
    log.Fatal(err)
  }
}

func main() {
  d := docline.New(&docline.Config{
    ResultsDirectory:    "./results",
    DefaultReportFormat: "html",
    DefaultTokenizer:    "space",
    DefaultCloneFinder:  "automatic",
  })

  automatic(d)
  heuristic(d)
  ngram(d)
  openai(d)
}