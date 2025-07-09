package apps

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/B9O2/canvas/containers"
	"github.com/B9O2/canvas/pixel"
	"github.com/B9O2/tabby"
)

//go:embed assets/RECIPE.sample.toml
var DefaultRecipe string

type InitRecipeApp struct {
	*tabby.BaseApplication
}

func (ma *InitRecipeApp) Detail() (string, string) {
	return "init", "Create a new config file (./RECIPE.toml)"
}

func (ma *InitRecipeApp) Main(args tabby.Arguments) (*tabby.TabbyContainer, error) {
	if args.Get("help").(bool) {
		name, desc := ma.Detail()
		ma.Help("[" + name + "] " + desc)
		return nil, nil
	}
	entrance := args.Get("entrance")
	if entrance == "" {
		entrance = "."
	}

	if _, err := os.Stat("RECIPE.toml"); err == nil {
		return nil, errors.New("RECIPE.toml already exists")
	}

	file, err := os.Create("RECIPE.toml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	text := fmt.Sprintf(DefaultRecipe, entrance)
	_, err = file.WriteString(text)
	if err != nil {
		return nil, err
	}

	//Canvas
	parts := strings.Split(text, "\n")
	height := uint(len(parts) + 4)

	body := containers.NewVStack(containers.NewAsciiArt(parts))
	body.SetFrame(0, 0, height, 0)
	body.SetHPadding(2)
	body.SetVPadding(1)
	body.SetBorder(pixel.Dot)

	vs := containers.NewVStack(containers.NewTextArea("RECIPE.toml"), body)
	tc := tabby.NewTabbyContainer(100, height+1, vs)

	return tc, nil
}

func NewInitRecipeApp() *InitRecipeApp {
	app := &InitRecipeApp{
		tabby.NewBaseApplication(false, nil),
	}
	app.SetParam("entrance", "", tabby.String("."), "e")
	app.SetParam("help", "Show help messages", tabby.Bool(false), "h")
	return app
}
