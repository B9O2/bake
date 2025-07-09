package main

import (
	"fmt"
	"os"

	"github.com/B9O2/bake/apps"

	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/B9O2/canvas/pixel"
	"github.com/B9O2/tabby"
)

func main() {
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	} else {
		args = []string{"default"}
	}

	buildApp := apps.NewBuildApp()
	initRecipeApp := apps.NewInitRecipeApp()
	listRecipesApp := apps.NewListRecipesApp()
	mainApp := apps.NewMainApp("main", "./RECIPE.toml", initRecipeApp, listRecipesApp)

	t := tabby.NewTabby("Bake", mainApp)
	t.SetUnknownApp(buildApp)

	tc, err := t.Run(args)
	if err != nil {
		Insp.Print(Error(err))
		return
	}
	if tc != nil {
		tc.Display(pixel.Space)
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
