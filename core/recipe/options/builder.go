package options

type OptionBuilder struct {
	Path string   `toml:"path"`
	Args []string `toml:"args"`
}

func (ob *OptionBuilder) Patch(patchOpt OptionBuilder) OptionBuilder {
	if patchOpt.Path != "" {
		ob.Path = patchOpt.Path
	}
	if len(patchOpt.Args) > 0 {
		ob.Args = patchOpt.Args
	}
	return *ob
}
