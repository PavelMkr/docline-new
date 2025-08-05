# Duplicate Finder

**Documentation Refactoring Toolkit** -- a clone finder and documentation refactoring tool within the DocLine project.

### Modes
- **Automatic mode** 
- **Interactive mode**
- **Ngram Duplicate Finder**
- **Heuristic Ngram Finder**

### Supported file formats
- DocBook (.xml, .dbk, .docbook)
- Microsoft Word (.doc, .docx)
- OpenDocument Text (.odt)
- Rich Text Format (.rtf)
- Markdown (.md)
- Plain Text (.txt)
- HTML (.html, .htm)

## Requirements

- **Go 1.23+**
- **Pandoc** - for converting documents to DocBook format

## Usage
```
go run .
```

### Linux:
**With GUI**: ```go run .```

**With CLI**: ```go run . -cli-... -input /path/to/file```

#### CLI params:

##### Modes:
- **cli-auto**: Run in automatic mode (CLI)
##### Other:
- **input**: Input file path
- **minClone**: Minimal clone length (tokens)
- **archetype**: Minimal archetype length (tokens) [auto mode]
- **strict**: Strict filtering [auto mode]
- **drl**: Convert to DRL [auto mode]
