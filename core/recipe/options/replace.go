package options

import (
	"regexp"

	"github.com/B9O2/filefinder"
)

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

func (orr *OptionReplace) ParseReplaceRule() (ReplaceRule, error) {
	r := ReplaceRule{
		DependencyReplace: orr.Dependency,
		ReplacementWords:  orr.Text,
	}

	var fileNameRegexps []*regexp.Regexp

	for _, fileNameRegexp := range orr.FileNameRegexps {
		if fileNameRegexp != "" {
			re, err := regexp.Compile(fileNameRegexp)
			if err != nil {
				return r, err
			}
			fileNameRegexps = append(fileNameRegexps, re)
		}
	}

	if len(orr.Dirs)+len(orr.FileNameRegexps) > 0 {
		r.Range = &filefinder.SearchRule{
			RuleName:        "OvO",
			DirRules:        orr.Dirs,
			FileNameRegexps: fileNameRegexps,
		}
	}

	return r, nil
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
