package config

import (
	"os"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Host           string   `yaml:"host"`
	Port           int      `yaml:"port"`
	AllowedOrigins []string `yaml:"allowedOrigins"`
}

type Database struct {
	Name           string `yaml:"name"`
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Migrate        bool   `yaml:"migrate"`
	MigrationsPath string `yaml:"migrationsPath"`
}

type Logging struct {
	Level                   string `yaml:"level"`
	Format                  string `yaml:"format"`
	SetReportCaller         bool   `yaml:"setReportCaller"`
	ArchiveLogs             bool   `yaml:"archiveLogs"`
	ArchiveBufferSize       int    `yaml:"archiveBufferSize"`
	ArchiveFrequencySeconds int    `yaml:"archiveFrequencySeconds"`
}

type Tracing struct {
	Enabled    bool    `yaml:"enabled"`
	SampleRate float64 `yaml:"sampleRate"`
}

type AWS struct {
	Endpoint  string `yaml:"endpoint"`
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
}

type Application struct {
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

func Load(config any) error {
	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		return errors.New("CONFIG_FILE env var not set")
	}

	return loadConfig(path, config)
}

func LoadFromFile(path string, config any) error {
	return loadConfig(path, config)
}

func loadConfig(path string, config any) error {
	if config == nil {
		return errors.New("config must not be nil")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "couldn't read config file from path: %s", path)
	}
	data = []byte(os.ExpandEnv(string(data)))

	if err := yaml.Unmarshal(data, config); err != nil {
		return errors.Wrapf(err, "couldn't unmarshal config file at path: %s", path)
	}

	return nil
}
