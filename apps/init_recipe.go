package apps

import (
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/b9o2/tabby"
)

//go:embed assets/RECIPE.sample.toml
var DefaultRecipe string

type InitRecipeApp struct {
	*tabby.BaseApplication
}

func (ma *InitRecipeApp) Detail() (string, string) {
	return "init", "Create a new config file (./RECIPE.toml)"
}

func (ma *InitRecipeApp) Main(args tabby.Arguments) error {
	if args.Get("help").(bool) {
		name, desc := ma.Detail()
		ma.Help("[" + name + "] " + desc)
		return nil
	}
	entrance := args.Get("entrance")
	if entrance == "" {
		entrance = "."
	}

	if _, err := os.Stat("RECIPE.toml"); err == nil {
		return errors.New("RECIPE.toml already exists")
	}

	file, err := os.Create("RECIPE.toml")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf(DefaultRecipe, entrance))
	if err != nil {
		return err
	}

	return nil
}

func NewInitRecipeApp() *InitRecipeApp {
	return &InitRecipeApp{
		tabby.NewBaseApplication(nil),
	}
}
