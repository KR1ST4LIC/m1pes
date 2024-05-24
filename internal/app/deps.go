package app

type initFunc func() error

func (a *App) InitDeps() {
	funcs := []initFunc{
		a.InitTelegramBot,
	}
	for _, f := range funcs {
		if err := f(); err != nil {
			panic(err)
		}
	}
}
