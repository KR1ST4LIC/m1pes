package config

import (
	"errors"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Bot    BotConfig    `yaml:"bot"`
	DBConn DBConnConfig `yaml:"db-conn"`
}

type BotConfig struct {
	Token string `yaml:"token"`
}

type DBConnConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func InitConfig() (*Config, error) {
	config := &Config{}

	cfgPath, err := getEnv("CONFIG_PATH")
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
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
