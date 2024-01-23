package main

import (
	"fmt"
	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/b9o2/tabby"
	"os"
)

func main() {
	defer func() {
		Insp.Print(Text("Finished", decorators.Magenta))
	}()
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	} else {
		args = []string{"default"}
	}

	ma := NewMainApp()
	ma.SetParam("list", "", tabby.BOOL, "l")
	ba := NewBuildApp()

	t := tabby.NewTabby("Bake", ma)
	t.SetUnknownApp(ba)
	err := t.Run(args)
	if err != nil {
		Insp.Print(Error(err))
		return
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
