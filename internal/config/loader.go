package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	// Global config instance
	cfg *Config

	// ConfigPath is the path to the config file
	ConfigPath string
)

// Init initializes the configuration system
func Init(configPath string) error {
	ConfigPath = configPath

	// Set defaults
	setDefaults()

	// Configure viper
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(DefaultConfigDir)
		viper.AddConfigPath(".")
	}

	// Environment variables
	viper.SetEnvPrefix("WTE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; use defaults
	}

	// Unmarshal into config struct
	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

// setDefaults sets default values in viper
func setDefaults() {
	// GOST defaults
	viper.SetDefault("gost.version", DefaultGOSTVersion)
	viper.SetDefault("gost.binary_path", DefaultGOSTBinaryPath)
	viper.SetDefault("gost.config_dir", DefaultGOSTConfigDir)
	viper.SetDefault("gost.config_file", DefaultGOSTConfigFile)

	// HTTP defaults
	viper.SetDefault("http.enabled", true)
	viper.SetDefault("http.port", DefaultHTTPPort)
	viper.SetDefault("http.auth.enabled", true)
	viper.SetDefault("http.auth.username", DefaultUsername)
	viper.SetDefault("http.auth.password", "")

	// HTTPS defaults
	viper.SetDefault("https.enabled", false)
	viper.SetDefault("https.port", DefaultHTTPSPort)
	viper.SetDefault("https.cert_path", DefaultGOSTConfigDir+"/cert.pem")
	viper.SetDefault("https.key_path", DefaultGOSTConfigDir+"/key.pem")
	viper.SetDefault("https.auth.enabled", true)
	viper.SetDefault("https.auth.username", DefaultUsername)
	viper.SetDefault("https.auth.password", "")

	// Shadowsocks defaults
	viper.SetDefault("shadowsocks.enabled", true)
	viper.SetDefault("shadowsocks.port", DefaultShadowsocksPort)
	viper.SetDefault("shadowsocks.method", DefaultShadowsocksMethod)
	viper.SetDefault("shadowsocks.password", "")

	// Firewall defaults
	viper.SetDefault("firewall.auto_configure", true)

	// Logging defaults
	viper.SetDefault("logging.level", DefaultLogLevel)
}

// Get returns the current configuration
func Get() *Config {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return cfg
}

// Set updates a configuration value
func Set(key string, value interface{}) error {
	viper.Set(key, value)

	// Re-unmarshal to update the config struct
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error updating config: %w", err)
	}

	return nil
}

// Save writes the current configuration to file
func Save() error {
	return SaveTo(WTEConfigFile)
}

// SaveTo writes the current configuration to a specific file
func SaveTo(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Load reads configuration from the specified file
func Load(path string) error {
	return Init(path)
}

// Reload reloads the configuration from the current file
func Reload() error {
	return Init(ConfigPath)
}

// Reset resets configuration to defaults
func Reset() {
	cfg = DefaultConfig()
}

// Exists checks if the config file exists
func Exists() bool {
	_, err := os.Stat(WTEConfigFile)
	return err == nil
}

// GetConfigPath returns the path to the active config file
func GetConfigPath() string {
	if viper.ConfigFileUsed() != "" {
		return viper.ConfigFileUsed()
	}
	return WTEConfigFile
}
