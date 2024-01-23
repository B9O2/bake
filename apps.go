package main

import (
	"bake/core"
	"bake/core/recipe"
	"fmt"
	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/b9o2/tabby"
)

type MainApp struct {
	*tabby.BaseApplication
}

func (ma *MainApp) Name() string {
	return "Bake"
}

func (ma *MainApp) Main(arguments tabby.Arguments) error {
	if show, err := arguments.Get("list"); err != nil {
		return err
	} else if show.(bool) {
		fmt.Println("[Show all recipes]")
		return nil
	}

	fmt.Println("[Other]")

	return nil
}

func NewMainApp() *MainApp {
	return &MainApp{
		tabby.NewBaseApplication(nil),
	}
}

type BuildApp struct {
	*tabby.BaseApplication
}

func (ba *BuildApp) BuildOne(pair recipe.BuildPair, cfg recipe.Config) error {
	b, err := core.NewGoProjectBuilder(".", pair.Builder.Path, false)
	if err != nil {
		return err
	}
	defer func() {
		if err = b.Close(); err != nil {
			Insp.Print(LEVEL_WARNING, Error(err), Path(b.ShadowPath()), Text("not clean"))
		} else {
			Insp.Print(LEVEL_INFO, Path(b.ShadowPath()), Text("cleaned"))
		}
	}()

	Insp.Print(Text("Shadow Project"), Path(b.ShadowPath()))
	if err = b.GoVendor(pair.Rule.DependencyReplace); err != nil {
		Insp.Print(Error(err))
		return err
	}
	if err = b.FileReplace(pair.Rule.ReplacementWords, pair.Rule.Range); err != nil {
		Insp.Print(Error(err))
		return err
	}
	realOutput, err := b.BuildProject(pair.Builder.Args, cfg.Entrance, cfg.Output, pair)
	if err != nil {
		return err
	} else {
		Insp.Print(Text("Build Successfully", decorators.Green), Text(realOutput))
	}
	return nil
}

func (ba *BuildApp) Name() string {
	return "Builder"
}

func (ba *BuildApp) Main(arguments tabby.Arguments) error {
	for _, r := range arguments.AppPath()[1:] { //跳过根应用
		Insp.Print(Text("Follow Recipe"), Text(r, decorators.Magenta))
		config, err := recipe.LoadConfig("./RECIPE.toml", r)
		if err != nil {
			return err
		}
		Insp.Print(Text("Entrance"), Text(config.Entrance, decorators.Blue))
		for _, pair := range config.Targets {
			Insp.Print(Text("Build Pair"), Text(pair.Tag(), decorators.Yellow), Text("<"+pair.Remote.Info()+">", decorators.Magenta))
			err = ba.BuildOne(pair, config)
			if err != nil {
				Insp.Print(Error(err))
				continue
			}
		}
	}
	return nil
}
func NewBuildApp() *BuildApp {
	return &BuildApp{
		tabby.NewBaseApplication(nil),
	}
}
