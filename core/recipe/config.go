package recipe

import (
	"bake/core/recipe/options"
	"bake/core/remotes"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"path/filepath"
)

type BuildPair struct {
	Platform string
	Arch     string
	Rule     options.ReplaceRule
	Remote   remotes.RemoteTarget
	Builder  options.OptionBuilder

	fileName string
}

func (bp BuildPair) Tag() string {
	return fmt.Sprintf("%s_%s", bp.Platform, bp.Arch)
}

func (bp BuildPair) Name() string {
	name := ""
	if bp.fileName != "" {
		name = bp.fileName
	} else {
		name = bp.Tag()
	}
	if bp.Platform == "windows" && filepath.Ext(name) != ".exe" {
		name += ".exe"
	}
	return name
}

type Config struct {
	Targets              []BuildPair
	DefaultBuilderOption options.OptionBuilder
	Entrance, Output     string
}

func LoadConfig(filePath, recipeName string) (Config, error) {
	doc := RecipeDoc{}
	if _, err := toml.DecodeFile(filePath, &doc); err != nil {
		return Config{}, err
	}

	if _, ok := doc.Recipes[recipeName]; !ok {
		return Config{}, errors.New("recipe '" + recipeName + "' not found")
	}
	//fmt.Println(doc.Recipes[recipeName])
	cfg, err := doc.Recipes[recipeName].ToConfig()
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
