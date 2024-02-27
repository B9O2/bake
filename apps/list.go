package apps

import (
	"bake/core/recipe"
	"fmt"

	"github.com/b9o2/tabby"
)

type ListRecipesApp struct {
	*tabby.BaseApplication
	ma *MainApp
}

func (lra *ListRecipesApp) Detail() (string, string) {
	return "ls", "Show all recipes"
}

func (lra *ListRecipesApp) Init(ma tabby.Application) error {
	lra.ma = ma.(*MainApp)
	return nil
}

func (lra *ListRecipesApp) Main(args tabby.Arguments) error {
	recipes, err := recipe.LoadAllRecipes(lra.ma.GetRecipePath())
	if err != nil {
		return err
	}
	fmt.Println("All Recipes:")
	for name, recipe := range recipes {
		fmt.Print("- ", name)
		if len(recipe.Desc) > 0 {
			fmt.Println(" '" + recipe.Desc + "'")
		} else {
			fmt.Println()
		}
	}
	return nil
}

func NewListRecipesApp(subApps ...tabby.Application) *ListRecipesApp {
	return &ListRecipesApp{
		tabby.NewBaseApplication(subApps),
		nil,
	}
}
