package apps

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/B9O2/bake/core"
	"github.com/B9O2/bake/core/recipe"
	"github.com/B9O2/bake/utils"

	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/B9O2/tabby"
)

type BuildApp struct {
	*tabby.BaseApplication
	ma *MainApp
}

func (ba *BuildApp) BuildOne(shadowBasePath string, pair recipe.BuildPair, cfg recipe.Config) error {
	if cfg.Debug {
		Insp.Print(LEVEL_INFO, Text("DEV MODE", decorators.Red))
	}

	b, err := core.NewGoProjectBuilder(shadowBasePath, ".", pair.Builder.Path, cfg.Debug)
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
	}
	Insp.Print(Text("Build Successfully", decorators.Green), Text(realOutput))

	//Zip
	if !pair.Output.Zip.IsEmpty() {
		source := filepath.Join(cfg.Output, pair.Output.Zip.Source)
		dest := filepath.Join(cfg.Output, pair.Output.Zip.Dest)
		Insp.Print(Text("Zipping Output", decorators.Yellow), Text(fmt.Sprintf("%s -> %s", source, dest), decorators.Magenta))
		if err = utils.Zip(source, dest, pair.Output.Zip.Password); err != nil {
			return err
		}
		Insp.Print(Text("Zipped Successfully", decorators.Green), Text(pair.Output.Zip.Dest, decorators.Magenta))
	}

	//SSH
	if !pair.Output.SSH.IsEmpty() {
		source := filepath.Join(cfg.Output, pair.Output.SSH.Source)
		Insp.Print(Text("SFTP", decorators.Yellow), Text(source, decorators.Magenta), Text("->", decorators.Yellow), Text(pair.Output.SSH.Dest, decorators.Magenta))
		client := utils.NewSSHClient(pair.Output.SSH.Host, pair.Output.SSH.Port, &utils.SSHAuthConfig{
			User:               pair.Output.SSH.User,
			Password:           pair.Output.SSH.Password,
			PrivateKeyPath:     pair.Output.SSH.PrivateKeyPath,
			PrivateKeyPassword: pair.Output.SSH.PrivateKeyPassword,
		})
		defer client.Close()

		err = client.Connect()
		if err != nil {
			return err
		}

		if utils.IsDir(source) {
			err = client.UploadDir(source, pair.Output.SSH.Dest)
			if err != nil {
				return err
			}
		} else {
			Insp.Print(Text("Uploading File", decorators.Yellow), Text(source, decorators.Magenta), Text("->", decorators.Yellow), Text(pair.Output.SSH.Dest, decorators.Magenta))
			err = client.UploadFile(source, pair.Output.SSH.Dest)
			if err != nil {
				return err
			}
		}
		Insp.Print(Text("SFTP Successfully", decorators.Green), Text(pair.Output.SSH.Dest, decorators.Magenta))
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

func (ba *BuildApp) Main(args tabby.Arguments) (*tabby.TabbyContainer, error) {
	shadowBasePath := path.Join(os.TempDir(), "BAKE_TMP")
	Insp.Print(Text("TempDir", decorators.Magenta), Path(shadowBasePath))
	for _, r := range args.AppPath()[1:] { //跳过根应用
		Insp.Print(Text("Follow Recipe"), Text(r, decorators.Magenta))
		config, err := recipe.LoadConfig(ba.ma.GetRecipePath(), r)
		if err != nil {
			return nil, err
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
	return nil, nil
}
func NewBuildApp() *BuildApp {
	return &BuildApp{
		tabby.NewBaseApplication(false, nil),
		nil,
	}
}
