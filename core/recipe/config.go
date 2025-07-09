package recipe

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/B9O2/bake/core/recipe/options"
	"github.com/B9O2/bake/core/targets"
	"github.com/B9O2/bake/utils"

	"github.com/BurntSushi/toml"
)

type BuildPair struct {
	Platform string
	Arch     string
	Rule     options.ReplaceRule
	Remote   targets.Target

	Builder options.OptionBuilder
	Output  options.OptionOutput
}

func (bp BuildPair) Tag() string {
	return fmt.Sprintf("%s_%s", bp.Platform, bp.Arch)
}

func (bp BuildPair) Name() string {
	name := ""
	if bp.Output.Path != "" {
		name = bp.Output.Path
	} else {
		name = bp.Tag()
	}
	if bp.Platform == "windows" && filepath.Ext(name) != ".exe" {
		name += ".exe"
	}
	return name
}

type Config struct {
	Debug            bool
	Targets          []BuildPair
	Entrance, Output string
}

func LoadAllRecipes(filePath string) (map[string]Recipe, error) {
	doc := RecipeDoc{}
	yes, err := utils.FileExists(filePath)
	if !yes {
		return map[string]Recipe{}, errors.New("Not a bake project, try 'bake init'")
	}
	if err != nil {
		return map[string]Recipe{}, err
	}
	if _, err := toml.DecodeFile(filePath, &doc); err != nil {
		return map[string]Recipe{}, err
	}
	return doc.Recipes, nil
}

func LoadConfig(filePath, recipeName string) (Config, error) {
	if recipes, err := LoadAllRecipes(filePath); err != nil {
		return Config{}, err
	} else {
		cfg, err := recipes[recipeName].ToConfig()
		if err != nil {
			return Config{}, err
		}
		return cfg, nil
	}
}
