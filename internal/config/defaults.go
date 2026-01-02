package config

const (
	// DefaultGOSTVersion is the default GOST version to install
	DefaultGOSTVersion = "3.0.0-rc10"

	// DefaultGOSTBinaryPath is where GOST binary is installed
	DefaultGOSTBinaryPath = "/usr/local/bin/gost"

	// DefaultConfigDir is the directory for WTE configuration
	DefaultConfigDir = "/etc/wte"

	// DefaultGOSTConfigDir is the directory for GOST configuration
	DefaultGOSTConfigDir = "/etc/gost"

	// DefaultGOSTConfigFile is the GOST configuration file path
	DefaultGOSTConfigFile = "/etc/gost/config.yaml"

	// DefaultHTTPPort is the default HTTP proxy port
	DefaultHTTPPort = 8080

	// DefaultHTTPSPort is the default HTTPS proxy port
	DefaultHTTPSPort = 8443

	// DefaultShadowsocksPort is the default Shadowsocks port
	DefaultShadowsocksPort = 9500

	// DefaultShadowsocksMethod is the default encryption method
	DefaultShadowsocksMethod = "aes-128-gcm"

	// DefaultUsername is the default proxy username
	DefaultUsername = "proxyuser"

	// DefaultLogLevel is the default logging level
	DefaultLogLevel = "info"

	// CredentialsFile is where credentials are saved
	CredentialsFile = "/root/proxy-credentials.txt"

	// SystemdServiceFile is the systemd service file path
	SystemdServiceFile = "/etc/systemd/system/gost.service"

	// WTEConfigFile is the main WTE configuration file
	WTEConfigFile = "/etc/wte/config.yaml"
)

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		GOST: GOSTConfig{
			Version:    DefaultGOSTVersion,
			BinaryPath: DefaultGOSTBinaryPath,
			ConfigDir:  DefaultGOSTConfigDir,
			ConfigFile: DefaultGOSTConfigFile,
		},
		HTTP: HTTPConfig{
			Enabled: true,
			Port:    DefaultHTTPPort,
			Auth: AuthConfig{
				Enabled:  true,
				Username: DefaultUsername,
				Password: "", // Will be auto-generated
			},
		},
		HTTPS: HTTPSConfig{
			Enabled:  false,
			Port:     DefaultHTTPSPort,
			CertPath: DefaultGOSTConfigDir + "/cert.pem",
			KeyPath:  DefaultGOSTConfigDir + "/key.pem",
			Auth: AuthConfig{
				Enabled:  true,
				Username: DefaultUsername,
				Password: "", // Will use same as HTTP
			},
		},
		Shadowsocks: ShadowsocksConfig{
			Enabled:  true,
			Port:     DefaultShadowsocksPort,
			Method:   DefaultShadowsocksMethod,
			Password: "", // Will be auto-generated
		},
		Firewall: FirewallConfig{
			AutoConfigure: true,
		},
		Logging: LoggingConfig{
			Level: DefaultLogLevel,
		},
	}
}
