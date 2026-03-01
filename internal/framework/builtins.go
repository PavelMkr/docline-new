package framework

// RegisterBuiltInPlugins registers core framework-level plugins (tokenizer,
// similarity calculator, filters) in the provided registry. Algorithms,
// document parsers/converters and report generators are registered from their
// own packages to avoid import cycles.
func RegisterBuiltInPlugins(reg *PluginRegistry) error {
	if err := reg.RegisterTextTokenizer(&SpaceTokenizer{}); err != nil {
		return err
	}
	if err := reg.RegisterSimilarityCalculator(&JaccardSimilarityCalculator{}); err != nil {
		return err
	}
	if err := reg.RegisterFilter(&StrictFilter{}); err != nil {
		return err
	}
	return nil
}

