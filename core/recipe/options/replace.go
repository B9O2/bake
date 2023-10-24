package options

import "github.com/B9O2/filefinder"

// OptionReplace 替换选项
type OptionReplace struct {
	Dependency      map[string]string `toml:"dependency"`
	Text            map[string]string `toml:"text"`
	Dirs            []string          `toml:"dir_rules"`
	FileNameRegexps []string          `toml:"file_regexps"`
}
type ReplaceRule struct {
	DependencyReplace map[string]string
	ReplacementWords  map[string]string
	Range             *filefinder.SearchRule
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
