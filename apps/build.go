package apps

import (
	"bake/core"
	"bake/core/recipe"
	"os"
	"path"

	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/b9o2/tabby"
)

type BuildApp struct {
	*tabby.BaseApplication
	ma *MainApp
}

func (ba *BuildApp) BuildOne(shadowBasePath string, pair recipe.BuildPair, cfg recipe.Config) error {
	b, err := core.NewGoProjectBuilder(shadowBasePath, ".", pair.Builder.Path, false)
	if err != nil {
		return err
	}
	defer func() {
		if err = b.Close(); err != nil {
			Insp.Print(LEVEL_WARNING, Error(err), Path(b.ShadowPath()), Text("not clean"))
		} else {
			//Insp.Print(LEVEL_INFO, Path(b.ShadowPath()), Text("cleaned"))
		}
	}()

	//Insp.Print(Text("Shadow Project"), Path(b.ShadowPath()))
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

func (ba *BuildApp) Detail() (string, string) {
	return "build", "Bake Builder"
}

func (ba *BuildApp) Init(ma tabby.Application) error {
	ba.ma = ma.(*MainApp)
	return nil
}

func (ba *BuildApp) Main(args tabby.Arguments) error {
	shadowBasePath := path.Join(os.TempDir(), "BAKE_TMP")
	Insp.Print(Text("TempDir", decorators.Magenta), Path(shadowBasePath))
	for _, r := range args.AppPath()[1:] { //跳过根应用
		Insp.Print(Text("Follow Recipe"), Text(r, decorators.Magenta))
		config, err := recipe.LoadConfig(ba.ma.GetRecipePath(), r)
		if err != nil {
			return err
		}
		Insp.Print(Text("Entrance"), Text(config.Entrance, decorators.Blue))
		for _, pair := range config.Targets {
			Insp.Print(Text("Build Pair"), Text(pair.Tag(), decorators.Yellow), Text("<"+pair.Remote.Info()+">", decorators.Magenta))
			err = ba.BuildOne(shadowBasePath, pair, config)
			if err != nil {
				Insp.Print(Error(err))
				continue
			}
		}
	}
	Insp.Print(Text("Finished", decorators.Magenta))
	return nil
}
func NewBuildApp() *BuildApp {
	return &BuildApp{
		tabby.NewBaseApplication(nil),
		nil,
	}
}
