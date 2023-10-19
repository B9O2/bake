package main

import (
	"bake/core"
	"fmt"
	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"os"
	"path/filepath"
)

func BuildOne(pair core.BuildPair, entrance, output string) error {
	b, err := core.NewGoProjectBuilder(".", "go", false)
	if err != nil {
		return err
	}
	defer func() {
		b.Close()
		Insp.Print(LEVEL_INFO, Path(b.ShadowPath()), Text("cleaned"))
	}()

	Insp.Print(Text("Shadow Project"), Path(b.ShadowPath()))
	output = filepath.Join(b.ProjectPath(), output)
	if err = b.GoVendor(pair.Rule.DependencyReplace); err != nil {
		Insp.Print(Error(err))
		return err
	}
	if err = b.FileReplace(pair.Rule.ReplacementWords, pair.Rule.Range); err != nil {
		Insp.Print(Error(err))
		return err
	}
	realOutput, err := b.BuildProject(entrance, output, pair)
	if err != nil {
		return err
	} else {
		Insp.Print(Text("Build Successfully", decorators.Green), Text(realOutput))
	}
	return nil
}

func main() {
	defer func() {
		Insp.Print(Text("Finished", decorators.Magenta))
	}()
	var recipes []string
	if len(os.Args) > 1 {
		recipes = os.Args[1:]
	} else {
		recipes = []string{"default"}
	}
	for _, recipe := range recipes {
		Insp.Print(Text("Follow Recipe"), Text(recipe, decorators.Magenta))
		config, err := core.LoadConfig("./RECIPE.toml", recipe)
		if err != nil {
			Insp.Print(Error(err))
			return
		}
		Insp.Print(Text("Entrance"), Text(config.Entrance, decorators.Blue))
		for _, pair := range config.Targets {
			Insp.Print(Text("Build Pair"), Text(pair.Tag(), decorators.Yellow), Text("<"+pair.Remote.Info()+">", decorators.Magenta))
			err = BuildOne(pair, config.Entrance, config.Output)
			if err != nil {
				Insp.Print(Error(err))
				continue
			}
		}
	}
}

func init() {
	Insp.SetTypeDecorations("_func", decorators.Invisible)
	Insp.NewAutoType("id", func() interface{} {
		return ":bake:"
	}, func(i interface{}) string {
		return fmt.Sprint(i)
	}, decorators.Gray)
	Insp.SetOrders("_time", Level, "id")
}