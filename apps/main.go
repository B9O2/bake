package apps

import (
	"github.com/b9o2/tabby"
)

type MainApp struct {
	*tabby.BaseApplication
	version    string
	recipePath string
}

func (ma *MainApp) Detail() (string, string) {
	return "bake", "Main App"
}

func (ma *MainApp) GetRecipePath() string {
	return ma.recipePath
}

func (ma *MainApp) GetVersion() string {
	return ma.version
}

// func (ma *MainApp) Init(tabby.Application) error {
// 	return nil
// }

func (ma *MainApp) Main(args tabby.Arguments) (*tabby.TabbyContainer, error) {
	if args.Get("help").(bool) {
		ma.Help("Bake" + " - " + ma.GetVersion())
	}
	return nil, nil
}

func NewMainApp(version, recipePath string, subApps ...tabby.Application) *MainApp {
	return &MainApp{
		tabby.NewBaseApplication(0, 0, subApps),
		version,
		recipePath,
	}
}
