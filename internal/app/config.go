package app

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Bot BotConfig `yaml:"bot"`
}

type BotConfig struct {
	Token string `yaml:"token"`
}

func (a *App) InitConfig() error {
	config := &Config{}

	cfgPath, err := getEnv("CONFIG_PATH")
	if err != nil {
		return err
	}

	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	a.cfg = config

	return nil
}

func getEnv(envKey string) (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("err loading: %v\n", err)
	}

	env := os.Getenv(envKey)
	if env == "" {
		return "", errors.New("missing address")
	}

	return env, nil
}
