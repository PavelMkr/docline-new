package framework

import (
	"fmt"
	"sync"
)

// PluginRegistry manages registration and lookup of plugins
type PluginRegistry struct {
	mu              sync.RWMutex
	cloneFinders    map[string]CloneFinder
	similarityCalcs map[string]SimilarityCalculator
	parsers         map[string]DocumentParser
	converters      map[string]DocumentConverter
	reportGenerators map[string]ReportGenerator
	tokenizers      map[string]TextTokenizer
	filters         map[string]Filter
	plugins         map[string]Plugin
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		cloneFinders:     make(map[string]CloneFinder),
		similarityCalcs: make(map[string]SimilarityCalculator),
		parsers:         make(map[string]DocumentParser),
		converters:      make(map[string]DocumentConverter),
		reportGenerators: make(map[string]ReportGenerator),
		tokenizers:      make(map[string]TextTokenizer),
		filters:         make(map[string]Filter),
		plugins:         make(map[string]Plugin),
	}
}

// RegisterCloneFinder registers a clone finder algorithm
func (r *PluginRegistry) RegisterCloneFinder(finder CloneFinder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := finder.Name()
	if _, exists := r.cloneFinders[name]; exists {
		return fmt.Errorf("clone finder '%s' already registered", name)
	}
	
	r.cloneFinders[name] = finder
	return nil
}

// GetCloneFinder retrieves a clone finder by name
func (r *PluginRegistry) GetCloneFinder(name string) (CloneFinder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	finder, exists := r.cloneFinders[name]
	if !exists {
		return nil, fmt.Errorf("clone finder '%s' not found", name)
	}
	
	return finder, nil
}

// ListCloneFinders returns all registered clone finders
func (r *PluginRegistry) ListCloneFinders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.cloneFinders))
	for name := range r.cloneFinders {
		names = append(names, name)
	}
	return names
}

// RegisterSimilarityCalculator registers a similarity calculator
func (r *PluginRegistry) RegisterSimilarityCalculator(calc SimilarityCalculator) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := calc.Name()
	if _, exists := r.similarityCalcs[name]; exists {
		return fmt.Errorf("similarity calculator '%s' already registered", name)
	}
	
	r.similarityCalcs[name] = calc
	return nil
}

// GetSimilarityCalculator retrieves a similarity calculator by name
func (r *PluginRegistry) GetSimilarityCalculator(name string) (SimilarityCalculator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	calc, exists := r.similarityCalcs[name]
	if !exists {
		return nil, fmt.Errorf("similarity calculator '%s' not found", name)
	}
	
	return calc, nil
}

// RegisterDocumentParser registers a document parser
func (r *PluginRegistry) RegisterDocumentParser(parser DocumentParser) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := parser.Name()
	if _, exists := r.parsers[name]; exists {
		return fmt.Errorf("document parser '%s' already registered", name)
	}
	
	r.parsers[name] = parser
	return nil
}

// GetDocumentParser retrieves a parser for the given file extension
func (r *PluginRegistry) GetDocumentParser(extension string) (DocumentParser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, parser := range r.parsers {
		for _, format := range parser.SupportedFormats() {
			if format == extension {
				return parser, nil
			}
		}
	}
	
	return nil, fmt.Errorf("no parser found for extension '%s'", extension)
}

// RegisterDocumentConverter registers a document converter
func (r *PluginRegistry) RegisterDocumentConverter(converter DocumentConverter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := converter.Name()
	if _, exists := r.converters[name]; exists {
		return fmt.Errorf("document converter '%s' already registered", name)
	}
	
	r.converters[name] = converter
	return nil
}

// GetDocumentConverter retrieves a converter by name
func (r *PluginRegistry) GetDocumentConverter(name string) (DocumentConverter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	converter, exists := r.converters[name]
	if !exists {
		return nil, fmt.Errorf("document converter '%s' not found", name)
	}
	
	return converter, nil
}

// RegisterReportGenerator registers a report generator
func (r *PluginRegistry) RegisterReportGenerator(generator ReportGenerator) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := generator.Name()
	if _, exists := r.reportGenerators[name]; exists {
		return fmt.Errorf("report generator '%s' already registered", name)
	}
	
	r.reportGenerators[name] = generator
	return nil
}

// GetReportGenerator retrieves a report generator by format
func (r *PluginRegistry) GetReportGenerator(format string) (ReportGenerator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, generator := range r.reportGenerators {
		if generator.Format() == format {
			return generator, nil
		}
	}
	
	return nil, fmt.Errorf("no report generator found for format '%s'", format)
}

// RegisterTextTokenizer registers a text tokenizer
func (r *PluginRegistry) RegisterTextTokenizer(tokenizer TextTokenizer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := tokenizer.Name()
	if _, exists := r.tokenizers[name]; exists {
		return fmt.Errorf("text tokenizer '%s' already registered", name)
	}
	
	r.tokenizers[name] = tokenizer
	return nil
}

// GetTextTokenizer retrieves a tokenizer by name
func (r *PluginRegistry) GetTextTokenizer(name string) (TextTokenizer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tokenizer, exists := r.tokenizers[name]
	if !exists {
		return nil, fmt.Errorf("text tokenizer '%s' not found", name)
	}
	
	return tokenizer, nil
}

// RegisterFilter registers a filter
func (r *PluginRegistry) RegisterFilter(filter Filter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := filter.Name()
	if _, exists := r.filters[name]; exists {
		return fmt.Errorf("filter '%s' already registered", name)
	}
	
	r.filters[name] = filter
	return nil
}

// GetFilter retrieves a filter by name
func (r *PluginRegistry) GetFilter(name string) (Filter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	filter, exists := r.filters[name]
	if !exists {
		return nil, fmt.Errorf("filter '%s' not found", name)
	}
	
	return filter, nil
}

// RegisterPlugin registers a generic plugin
func (r *PluginRegistry) RegisterPlugin(plugin Plugin, config map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' already registered", name)
	}
	
	if err := plugin.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize plugin '%s': %v", name, err)
	}
	
	r.plugins[name] = plugin
	return nil
}

// UnregisterPlugin unregisters a plugin
func (r *PluginRegistry) UnregisterPlugin(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	plugin, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}
	
	if err := plugin.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown plugin '%s': %v", name, err)
	}
	
	delete(r.plugins, name)
	return nil
}



