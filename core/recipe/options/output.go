package options

type OptionOutput struct {
	Path string          `toml:"path"`
	Zip  OptionZIP       `toml:"zip"`
	SSH  OptionSSHOutput `toml:"ssh"`
}

func (oo *OptionOutput) Patch(patchOp OptionOutput) OptionOutput {
	if patchOp.Path != "" {
		oo.Path = patchOp.Path
	}
	oo.Zip = oo.Zip.Patch(patchOp.Zip)
	oo.SSH = oo.SSH.Patch(patchOp.SSH)
	return *oo
}

type OptionZIP struct {
	Source   string `toml:"source"`
	Dest     string `toml:"dest"`
	Password string `toml:"password"`
}

func (oz *OptionZIP) Patch(patchOz OptionZIP) OptionZIP {
	if patchOz.Source != "" {
		oz.Source = patchOz.Source
	}
	if patchOz.Dest != "" {
		oz.Dest = patchOz.Dest
	}
	if patchOz.Password != "" {
		oz.Password = patchOz.Password
	}
	return *oz
}

func (oz *OptionZIP) IsEmpty() bool {
	return oz.Source == "" && oz.Dest == ""
}

type OptionSSHOutput struct {
	OptionSSH
	Source string `toml:"source"`
	Dest   string `toml:"dest"`
}

func (oso *OptionSSHOutput) Patch(patchOpt OptionSSHOutput) OptionSSHOutput {
	oso.OptionSSH = oso.OptionSSH.Patch(patchOpt.OptionSSH)
	if patchOpt.Source != "" {
		oso.Source = patchOpt.Source
	}
	if patchOpt.Dest != "" {
		oso.Dest = patchOpt.Dest
	}
	return *oso
}

func (oso *OptionSSHOutput) IsEmpty() bool {
	return oso.Source == "" && oso.Dest == ""
}
