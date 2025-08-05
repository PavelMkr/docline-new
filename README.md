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

### Linux:
**With GUI**: ```go run .```

**With CLI**: ```go run . -cli-... ...```

#### CLI params:

- `cli-auto`: Run in automatic mode (CLI) +
    - `minClone`: Minimal clone length (tokens);
    - `drl`: Convert to DRL;
    - `archetype`: Minimal archetype length (tokens);
    - `strict`: Strict filtering;
    - `input`: Input file path;

- `cli-interactive`: Run in interactive mode (CLI) *
    - `minClone`: Minimal clone length (tokens);
    - `maxClone`: Maximal clone length (tokens);
    - `minGroup`: Minimal Group Power (number of clones);
    - `use-archetype`: Archetype calculation;
    - `input`: Input file path;

- `cli-ngram`: Run in ngram duplicate mode
    - `minClone`: Minimal clone length (tokens);
    - `maxDist`: Maximal edit distance (Levenshtein);
    - `maxEdit`: Maximal fuzzy hash distance;
    - `sourceLang`: Source document language;
    - `input`: Input file path;

- `cli-heuristic`: 
    - `extention`: Extension point values;
    - `input`: Input file path;
