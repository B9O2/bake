package options

// Options 每对平台架构的具体设置
type Options struct {
	Output      OptionOutput  `toml:"output"`
	ReplaceRule OptionReplace `toml:"replace"`
	Docker      OptionDocker  `toml:"docker"`
}

// Patch 对之前的选项进行补充
func (opt *Options) Patch(patchOpt Options) Options {
	opt.ReplaceRule = opt.ReplaceRule.Patch(patchOpt.ReplaceRule)
	opt.Docker = opt.Docker.Patch(patchOpt.Docker)
	opt.Output = opt.Output.Patch(patchOpt.Output)
	return *opt
}
