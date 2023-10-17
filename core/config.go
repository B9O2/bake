package core

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
)

type BuildTarget struct {
	Platform string
	Arch     string
	Rule     ReplaceRule
}

func (bt BuildTarget) Tag() string {
	return fmt.Sprintf("%s_%s", bt.Platform, bt.Arch)
}

type Config struct {
	Targets  []BuildTarget
	Entrance string
	Output   string
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
