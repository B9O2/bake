package recipe

import (
	_ "embed"
	"errors"
	"strings"

	"github.com/B9O2/bake/core/recipe/options"
	"github.com/B9O2/bake/core/targets"
	"github.com/B9O2/bake/utils"
)

//go:embed assets/PAIRS
var AllPairs string

type ArchOption map[string]options.Options

func (p ArchOption) Range(f func(arch string, option options.Options) bool) {
	for arch, option := range p {
		if arch == "all_arch" {
			continue
		}
		if !f(arch, option) {
			break
		}
	}
}
func (p ArchOption) AllArchOption() options.Options {
	return p["all_arch"]
}

type Recipe struct {
	Debug       bool       `toml:"debug"`
	Desc        string     `toml:"desc"`
	Entrance    string     `toml:"entrance"`
	Output      string     `toml:"output"`
	Pairs       []string   `toml:"pairs"`
	AllPlatform ArchOption `toml:"all_platform"`
	Darwin      ArchOption `toml:"darwin"`
	Linux       ArchOption `toml:"linux"`
	Windows     ArchOption `toml:"windows"`
}

func (r Recipe) ToConfig() (Config, error) {
	cfg := Config{
		Debug:  r.Debug,
		Output: "bake_bin",
	}
	mid := map[string]map[string]options.Options{}
	if len(r.Pairs) <= 0 {
		r.Pairs = strings.Split(AllPairs, "\n")
	}
	//在中间结构mid中初始化所有目标平台架构
	for _, pair := range r.Pairs {
		if platform, arch, ok := strings.Cut(pair, "/"); ok {
			if _, ok = mid[platform]; !ok {
				mid[platform] = map[string]options.Options{}
			}
			mid[platform][arch] = options.Options{}
		}
	}

	PatchOption := func(platform string, ao ArchOption) {
		if _, ok := mid[platform]; !ok {
			return
		}
		opt := ao.AllArchOption()
		for platform, archOption := range mid {
			for midArch, midOption := range archOption {
				mid[platform][midArch] = midOption.Patch(opt)
			}
		}
		ao.Range(func(arch string, option options.Options) bool {
			if _, ok := mid[platform][arch]; ok {
				midOption := mid[platform][arch]
				mid[platform][arch] = midOption.Patch(option)
			}
			return true
		})
	}

	//遍历所有目标平台架构，应用全平台设置
	for platform := range mid {
		PatchOption(platform, r.AllPlatform)
	}

	//分别应用特殊平台设置
	PatchOption("darwin", r.Darwin)
	PatchOption("linux", r.Linux)
	PatchOption("windows", r.Windows)

	for platform, archOption := range mid {
		for arch, option := range archOption {
			rr, err := option.ReplaceRule.ParseReplaceRule()
			if err != nil {
				return cfg, err
			}

			bp := BuildPair{
				Platform: platform,
				Arch:     arch,
				Rule:     rr,
				Remote:   targets.NewLocalTarget(platform, arch), //默认本地编译
				Builder: options.OptionBuilder{
					Path: "go",
					Args: []string{
						"-trimpath",
						"-ldflags",
						"-w -s",
					},
					Env: map[string]string{},
				},
			}

			bp.Output.Patch(option.Output)
			bp.Builder.Patch(option.Builder)

			//配置了Docker目标
			if option.Docker.Host != "" {
				bp.Remote = targets.NewDockerTarget(option.Docker.Host,
					option.Docker.Container,
					option.Docker.Image,
					option.Docker.Temp,
					platform,
					arch)
			}

			//配置了SSH目标
			if option.SSH.Host != "" {
				sshCfg := &utils.SSHAuthConfig{
					User:               option.SSH.User,
					Password:           option.SSH.Password,
					PrivateKeyPath:     option.SSH.PrivateKeyPath,
					PrivateKeyPassword: option.SSH.PrivateKeyPassword,
				}

				bp.Remote = targets.NewSSHTargetWithConfig(
					option.SSH.Host,
					option.SSH.Port,
					option.SSH.Temp,
					platform,
					arch,
					sshCfg,
				)
			}

			cfg.Targets = append(cfg.Targets, bp)
		}
	}

	if r.Entrance == "" {
		return cfg, errors.New("no entrance")
	} else {
		cfg.Entrance = r.Entrance
	}

	if r.Output != "" {
		cfg.Output = r.Output
	}

	return cfg, nil
}

type RecipeDoc struct {
	Recipes map[string]Recipe `toml:"recipes"`
}
