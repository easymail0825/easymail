package database

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type App struct {
	Name      string            `yaml:"name"`
	Family    string            `yaml:"family"`
	Listen    string            `yaml:"listen"`
	Enable    bool              `yaml:"enable"`
	Parameter map[string]string `yaml:"parameter"`
}

type AppConfig struct {
	Mysql   string `yaml:"mysql"`
	Redis   string `yaml:"redis"`
	LogFile string `yaml:"log_file"`
	Apps    []App  `yaml:"apps"`
}

func ReadAppConfig(filename string) (*AppConfig, error) {
	appConfig := AppConfig{}

	// read config file
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Println("Error reading file:", err)
		return nil, err
	}
	err = yaml.Unmarshal(data, &appConfig)
	if err != nil {
		return nil, err
	}
	return &appConfig, nil
}
