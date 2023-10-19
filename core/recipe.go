package core

import (
	"bake/core/remotes"
	_ "embed"
	"errors"
	//. "github.com/B9O2/Inspector/templates/simple"
	"github.com/B9O2/filefinder"
	"strings"
)

//go:embed static/PAIRS
var AllPairs string

type OptionReplace struct {
	Dependency      map[string]string `toml:"dependency"`
	Text            map[string]string `toml:"text"`
	Dirs            []string          `toml:"dir_rules"`
	FileNameRegexps []string          `toml:"file_regexps"`
}

func (orr *OptionReplace) ParseReplaceRule() ReplaceRule {
	r := ReplaceRule{
		DependencyReplace: orr.Dependency,
		ReplacementWords:  orr.Text,
	}

	if len(orr.Dirs)+len(orr.FileNameRegexps) > 0 {
		r.Range = &filefinder.SearchRule{
			RuleName:        "OvO",
			DirRules:        orr.Dirs,
			FileNameRegexps: orr.FileNameRegexps,
		}
	}

	return r
}
func (orr *OptionReplace) Patch(por OptionReplace) OptionReplace {
	if orr.Dependency == nil {
		orr.Dependency = map[string]string{}
	}
	for k, v := range por.Dependency {
		orr.Dependency[k] = v
	}

	if orr.Text == nil {
		orr.Text = map[string]string{}
	}
	for k, v := range por.Text {
		orr.Text[k] = v
	}
	return *orr
}

type OptionDocker struct {
	Host      string `toml:"host"`
	Container string `toml:"container"`
	Temp      string `toml:"temp""`
}

func (od *OptionDocker) Patch(patchOpt OptionDocker) OptionDocker {
	if patchOpt.Host != "" {
		od.Host = patchOpt.Host
	}
	if patchOpt.Container != "" {
		od.Container = patchOpt.Container
	}
	if patchOpt.Temp != "" {
		od.Temp = patchOpt.Temp
	}
	return *od
}

// Options 每对平台架构的具体设置
type Options struct {
	Name        string        //未启用
	ReplaceRule OptionReplace `toml:"replace"`
	Docker      OptionDocker  `toml:"docker"`
}

// Patch 对之前的选项进行补充
func (opt *Options) Patch(patchOpt Options) Options {
	opt.ReplaceRule = opt.ReplaceRule.Patch(patchOpt.ReplaceRule)
	opt.Docker = opt.Docker.Patch(patchOpt.Docker)
	return *opt
}

type ArchOption map[string]Options

func (p ArchOption) Range(f func(arch string, option Options) bool) {
	for arch, option := range p {
		if arch == "all_arch" {
			continue
		}
		if !f(arch, option) {
			break
		}
	}
}
func (p ArchOption) AllArchOption() Options {
	option, _ := p["all_arch"]
	return option
}

type Recipe struct {
	Entrance    string     `toml:"entrance"`
	Output      string     `toml:"output"`
	Pairs       []string   `toml:"pairs"`
	AllPlatform ArchOption `toml:"all_platform"`
	Darwin      ArchOption `toml:"darwin"`
	Linux       ArchOption `toml:"linux"`
	Windows     ArchOption `toml:"windows"`
}

func (r Recipe) ToConfig() (Config, error) {
	cfg := Config{}
	mid := map[string]map[string]Options{}
	if len(r.Pairs) <= 0 {
		r.Pairs = strings.Split(AllPairs, "\n")
	}
	//在中间结构mid中初始化所有目标平台架构
	for _, pair := range r.Pairs {
		if platform, arch, ok := strings.Cut(pair, "/"); ok {
			if _, ok = mid[platform]; !ok {
				mid[platform] = map[string]Options{}
			}
			mid[platform][arch] = Options{}
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
		ao.Range(func(arch string, option Options) bool {
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
				Rule:     option.ReplaceRule.ParseReplaceRule(),
				Remote:   remotes.NewLocalTarget(platform, arch), //默认本地编译
			}

			if option.Docker.Host != "" {
				bt.Remote = remotes.NewDockerTarget(option.Docker.Host,
					option.Docker.Container,
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

	if r.Output == "" {
		cfg.Output = "bake_out"
	} else {
		cfg.Output = r.Output
	}
	return cfg, nil
}

type RecipeDoc struct {
	Recipes map[string]Recipe `toml:"recipes"`
}
