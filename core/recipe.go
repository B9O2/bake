package core

import (
	_ "embed"
	"fmt"
	"gitlab.huaun.com/lr/filefinder"
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

type Options struct {
	ReplaceRule OptionReplace `toml:"replace"`
}

func (opt *Options) Patch(patchOpt Options) Options {
	opt.ReplaceRule = opt.ReplaceRule.Patch(patchOpt.ReplaceRule)
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

		fmt.Println("Platform <", platform, ">Option:", ao)

		ao.Range(func(arch string, option Options) bool {
			if _, ok := mid[platform][arch]; ok {
				midOption := mid[platform][arch]
				mid[platform][arch] = midOption.Patch(option)
			}
			return true
		})
	}

	for platform := range mid {
		PatchOption(platform, r.AllPlatform)
	}

	PatchOption("darwin", r.Darwin)
	PatchOption("linux", r.Linux)
	PatchOption("windows", r.Windows)
	for platform, archOption := range mid {
		for arch, option := range archOption {
			cfg.Targets = append(cfg.Targets, BuildTarget{
				Platform: platform,
				Arch:     arch,
				Entrance: r.Entrance,
				Rule:     option.ReplaceRule.ParseReplaceRule(),
			})
		}
	}

	return cfg, nil
}

type RecipeDoc struct {
	Recipes map[string]Recipe `toml:"recipes"`
}
