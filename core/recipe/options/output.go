package options

type OptionOutput struct {
	Path string `toml:"path"`
}

func (op OptionOutput) Patch(patchOp OptionOutput) OptionOutput {
	if patchOp.Path != "" {
		op.Path = patchOp.Path
	}
	return op
}
