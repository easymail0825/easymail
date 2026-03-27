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

	InitDB bool `yaml:"init_db"`

	Network NetworkConfig `yaml:"network"`

	Feature FeatureConfig `yaml:"feature"`

	Postfix PostfixConfig `yaml:"postfix"`

	LMTP LMTPConfig `yaml:"lmtp"`

	Observability ObservabilityConfig `yaml:"observability"`

	Apps []App `yaml:"apps"`
}

type NetworkConfig struct {
	DNS NetworkDNSConfig `yaml:"dns"`
}

type NetworkDNSConfig struct {
	NameServer string `yaml:"nameserver"`
}

type FeatureConfig struct {
	IP FeatureIPConfig `yaml:"ip"`
}

type FeatureIPConfig struct {
	Region         bool   `yaml:"region"`
	RegionCityMMDB string `yaml:"region_city_mmdb"`
}

type PostfixConfig struct {
	Execute PostfixExecuteConfig `yaml:"execute"`
	Log     PostfixLogConfig     `yaml:"log"`
	Sync    PostfixSyncConfig    `yaml:"sync"`
}

type PostfixExecuteConfig struct {
	Postconf  string `yaml:"postconf"`
	Postqueue string `yaml:"postqueue"`
	Postcat   string `yaml:"postcat"`
	Postsuper string `yaml:"postsuper"`
	Postmap   string `yaml:"postmap"`
	Postfix   string `yaml:"postfix"`
}

type PostfixLogConfig struct {
	Mail string `yaml:"mail"`
}

type PostfixSyncConfig struct {
	VirtualMailboxDomains string `yaml:"virtual_mailbox_domains"`
}

type StorageConfig struct {
	Root string `yaml:"root"`
	Data string `yaml:"data"`
}

type LMTPConfig struct {
	Storage StorageConfig `yaml:"storage"`
}

type ObservabilityConfig struct {
	SessionTrace SessionTraceConfig `yaml:"session_trace"`
}

type SessionTraceConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Sink      string `yaml:"sink"` // file|db
	FilePath  string `yaml:"file_path"`
	QueueSize int    `yaml:"queue_size"`
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
