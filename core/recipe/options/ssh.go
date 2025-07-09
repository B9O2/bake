package options

// OptionSSH SSH选项
type OptionSSH struct {
	User               string `toml:"user"`
	Host               string `toml:"host"`
	Port               int    `toml:"port"`
	Password           string `toml:"password"`
	PrivateKeyPath     string `toml:"private_key_path"`
	PrivateKeyPassword string `toml:"private_key_password"`
}

func (os *OptionSSH) Patch(patchOpt OptionSSH) OptionSSH {
	if patchOpt.Host != "" {
		os.Host = patchOpt.Host
	}

	if patchOpt.Port != 0 {
		os.Port = patchOpt.Port
	}

	if patchOpt.User != "" {
		os.User = patchOpt.User
	}

	if patchOpt.Password != "" {
		os.Password = patchOpt.Password
	}

	if patchOpt.PrivateKeyPath != "" {
		os.PrivateKeyPath = patchOpt.PrivateKeyPath
	}

	if patchOpt.PrivateKeyPassword != "" {
		os.PrivateKeyPassword = patchOpt.PrivateKeyPassword
	}
	return *os
}

type OptionSSHBuild struct {
	OptionSSH
	Temp string `toml:"temp"`
}

func (osb *OptionSSHBuild) Patch(patchOpt OptionSSHBuild) OptionSSHBuild {
	osb.OptionSSH = osb.OptionSSH.Patch(patchOpt.OptionSSH)
	if patchOpt.Temp != "" {
		osb.Temp = patchOpt.Temp
	}
	return *osb
}

