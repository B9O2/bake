package recipe

import (
	"bake/core/recipe/options"
	"bake/core/remotes"
	_ "embed"
	"errors"
	"strings"
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
			bt := BuildPair{
				Platform: platform,
				Arch:     arch,
				fileName: option.Output.Path,
				Rule:     option.ReplaceRule.ParseReplaceRule(),
				Remote:   remotes.NewLocalTarget(platform, arch), //默认本地编译
				Builder: options.OptionBuilder{
					Path: "go",
					Args: []string{
						"-trimpath",
						"-ldflags",
						"-w -s",
					},
				},
			}

			bt.Builder.Patch(option.Builder)

			if option.Docker.Host != "" {
				bt.Remote = remotes.NewDockerTarget(option.Docker.Host,
					option.Docker.Container,
					option.Docker.Image,
					option.Docker.Temp,
					platform,
					arch)
			}

			cfg.Targets = append(cfg.Targets, bt)
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
