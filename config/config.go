package config

import (
	"os"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"
)

// Server represents the server configuration with host, port, and allowed origins.
type Server struct {
	Host           string   `yaml:"host"`
	Port           int      `yaml:"port"`
	AllowedOrigins []string `yaml:"allowedOrigins"`
}

// Database represents the configuration details for a database connection, including credentials and migration options.
type Database struct {
	Name           string `yaml:"name"`
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Migrate        bool   `yaml:"migrate"`
	MigrationsPath string `yaml:"migrationsPath"`
}

// Logging represents the logging configuration including level, format, caller reporting, and log archiving settings.
// Level specifies the logging verbosity level (e.g., debug, info, warn, error).
// Format defines the output format of logs (e.g., JSON, plain text).
// SetReportCaller determines if log entries include the calling method or file information.
// ArchiveLogs indicates whether logs should be archived.
// ArchiveBufferSize is the size of the buffer for storing logs in memory before archiving.
// ArchiveFrequencySeconds defines the interval in seconds between successive log archiving actions.
type Logging struct {
	Level                   string `yaml:"level"`
	Format                  string `yaml:"format"`
	SetReportCaller         bool   `yaml:"setReportCaller"`
	ArchiveLogs             bool   `yaml:"archiveLogs"`
	ArchiveBufferSize       int    `yaml:"archiveBufferSize"`
	ArchiveFrequencySeconds int    `yaml:"archiveFrequencySeconds"`
}

// Tracing represents the tracing configuration for distributed system monitoring and sampling settings.
// Enabled specifies whether tracing is active.
// SampleRate determines the fraction of requests to sample for tracing.
type Tracing struct {
	Enabled    bool    `yaml:"enabled"`
	SampleRate float64 `yaml:"sampleRate"`
}

type Monitoring struct {
	Tracing        Tracing `yaml:"tracing"`
	LogsEnabled    bool    `yaml:"logsEnabled"`
	MetricsEnabled bool    `yaml:"metricsEnabled"`
}

// AWS represents the configuration required to interact with AWS services, including credentials and resource details.
type AWS struct {
	Endpoint  string `yaml:"endpoint"`
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
	AccessKey string `yaml:"accessKeyId"`
	SecretKey string `yaml:"secretAccessKey"`
}

// Application represents the configuration details of the application, including its version and deployment environment.
type Application struct {
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

// Load retrieves configuration data from a file specified by the CONFIG_FILE environment variable and unmarshals it.
// Returns an error if the environment variable is unset, the file cannot be read, or unmarshalling fails.
func Load(config any) error {
	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		return errors.New("CONFIG_FILE env var not set")
	}

	return loadConfig(path, config)
}

// LoadFromFile reads a configuration file from the specified path and unmarshals its content into the provided config.
// Returns an error if the file cannot be read, the config is nil, or unmarshalling fails.
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
