package options

type OptionBuilder struct {
	Path string            `toml:"path"`
	Args []string          `toml:"args"`
	Env  map[string]string `toml:"env"`
}

func (ob *OptionBuilder) Patch(patchOpt OptionBuilder) OptionBuilder {
	if patchOpt.Path != "" {
		ob.Path = patchOpt.Path
	}
	if len(patchOpt.Args) > 0 {
		ob.Args = patchOpt.Args
	}

	if ob.Env == nil {
		ob.Env = map[string]string{}
	}
	for k, v := range patchOpt.Env {
		ob.Env[k] = v
	}
	return *ob
}
