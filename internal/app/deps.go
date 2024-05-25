package app

type initFunc func() error

func (a *App) InitDeps() error {
	funcs := []initFunc{
		a.InitLogging,
		a.InitConfig,
		a.InitTelegramBot,
	}
	for _, f := range funcs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}
