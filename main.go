package main

import (
	"ShadowProject/core"
	"fmt"
	"os"
	"path/filepath"
)

func BuildOne(target core.BuildTarget) error {
	b, err := core.NewBuilder(".", "go")
	if err != nil {
		return err
	}
	defer func() {
		b.Close()
		fmt.Println("bake: Temp cleaned")
	}()
	fmt.Println("[v]Shadow Project has been generated at '" + b.ShadowPath() + "'")
	out := filepath.Join(b.ProjectPath(), "bake_bin", target.Tag())
	if target.Output != "" {
		out = target.Output
	}
	if err := b.GoVendor(target.Rule); err != nil {
		return err
	}
	err = b.BuildProject(target.Entrance, target.Platform, target.Arch, out)
	if err != nil {
		return err
	} else {
		fmt.Println("[v]Build Successfully:", out)
	}
	return nil
}

func main() {
	defer func() {
		fmt.Println("bake: Finished")
	}()
	var recipes []string
	if len(os.Args) > 1 {
		recipes = os.Args[1 : len(os.Args)-1]
	} else {
		recipes = []string{"default"}
	}
	for _, recipe := range recipes {
		config, err := core.LoadConfig("./RECIPE.toml", recipe)
		if err != nil {
			fmt.Println("bake:" + err.Error())
			return
		}

		fmt.Println("TARGETS:", config.Targets)

		for _, target := range config.Targets {
			fmt.Println("NowTarget:", target)
			err := BuildOne(target)
			if err != nil {
				fmt.Println("[x]Build Error:", err)
				continue
			}
		}
	}
}
