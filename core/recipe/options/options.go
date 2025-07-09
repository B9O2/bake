package options

// Options 每对平台架构的具体设置
type Options struct {
	Builder     OptionBuilder  `toml:"builder"`
	Output      OptionOutput   `toml:"output"`
	ReplaceRule OptionReplace  `toml:"replace"`
	Docker      OptionDocker   `toml:"docker"`
	SSH         OptionSSHBuild `toml:"ssh"`
}

// Patch 对之前的选项进行补充
func (opt *Options) Patch(patchOpt Options) Options {
	opt.ReplaceRule = opt.ReplaceRule.Patch(patchOpt.ReplaceRule)
	opt.Docker = opt.Docker.Patch(patchOpt.Docker)
	opt.SSH = opt.SSH.Patch(patchOpt.SSH)
	opt.Output = opt.Output.Patch(patchOpt.Output)
	opt.Builder = opt.Builder.Patch(patchOpt.Builder)
	return *opt
}
