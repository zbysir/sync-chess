package conf

var StdApp App

type App struct {
	RunModel string
}

func init() {
	StdApp = App{
		RunModel: "dev",
	}
}
