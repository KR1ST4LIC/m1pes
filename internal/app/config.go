package app

import "m1pes/internal/config"

func (a *App) InitConfig() error {
	cfg, err := config.InitConfig()
	if err != nil {
		return err
	}
	a.cfg = cfg

	return nil
}
