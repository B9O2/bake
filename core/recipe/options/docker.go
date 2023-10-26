package options

// OptionDocker Docker选项
type OptionDocker struct {
	Host      string `toml:"host"`
	Container string `toml:"container"`
	Image     string `toml:"image"`
	Temp      string `toml:"temp"`
}

func (od *OptionDocker) Patch(patchOpt OptionDocker) OptionDocker {
	if patchOpt.Host != "" {
		od.Host = patchOpt.Host
	}
	if patchOpt.Container != "" {
		od.Container = patchOpt.Container
	}
	if patchOpt.Image != "" {
		od.Image = patchOpt.Image
	}
	if patchOpt.Temp != "" {
		od.Temp = patchOpt.Temp
	}
	return *od
}
