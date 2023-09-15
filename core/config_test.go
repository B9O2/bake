package core

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("/Users/restr0/Projects/HuaunProjects/botDefinger/RECIPE.toml", "default")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("\n\nTargets:")
		for _, t := range config.Targets {
			fmt.Println(t.Tag())
			fmt.Println(t.Rule)
			fmt.Println(t.Entrance)
		}
	}
}

func TestLoadRecipeDoc(t *testing.T) {
	doc := RecipeDoc{}
	_, err := toml.DecodeFile("/Users/restr0/Projects/HuaunProjects/botDefinger/RECIPE.toml", &doc)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(doc)
		fmt.Println("\n\nTargets:")
		/*
			for name, recipe := range doc.Recipes {
				for _, t := range recipe.Target {
					fmt.Println(name, "::", t)
				}
			}

		*/

	}
}
