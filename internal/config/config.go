package config

// Config represents the main application configuration
type Config struct {
	GOST        GOSTConfig        `yaml:"gost" mapstructure:"gost"`
	HTTP        HTTPConfig        `yaml:"http" mapstructure:"http"`
	HTTPS       HTTPSConfig       `yaml:"https" mapstructure:"https"`
	Shadowsocks ShadowsocksConfig `yaml:"shadowsocks" mapstructure:"shadowsocks"`
	Firewall    FirewallConfig    `yaml:"firewall" mapstructure:"firewall"`
	Logging     LoggingConfig     `yaml:"logging" mapstructure:"logging"`
}

// GOSTConfig holds GOST binary configuration
type GOSTConfig struct {
	Version    string `yaml:"version" mapstructure:"version"`
	BinaryPath string `yaml:"binary_path" mapstructure:"binary_path"`
	ConfigDir  string `yaml:"config_dir" mapstructure:"config_dir"`
	ConfigFile string `yaml:"config_file" mapstructure:"config_file"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
}

// HTTPConfig holds HTTP proxy configuration
type HTTPConfig struct {
	Enabled bool       `yaml:"enabled" mapstructure:"enabled"`
	Port    int        `yaml:"port" mapstructure:"port"`
	Auth    AuthConfig `yaml:"auth" mapstructure:"auth"`
}

// HTTPSConfig holds HTTPS proxy configuration
type HTTPSConfig struct {
	Enabled  bool       `yaml:"enabled" mapstructure:"enabled"`
	Port     int        `yaml:"port" mapstructure:"port"`
	CertPath string     `yaml:"cert_path" mapstructure:"cert_path"`
	KeyPath  string     `yaml:"key_path" mapstructure:"key_path"`
	Auth     AuthConfig `yaml:"auth" mapstructure:"auth"`
}

// ShadowsocksConfig holds Shadowsocks configuration
type ShadowsocksConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Method   string `yaml:"method" mapstructure:"method"`
	Password string `yaml:"password" mapstructure:"password"`
}

// FirewallConfig holds firewall configuration
type FirewallConfig struct {
	AutoConfigure bool `yaml:"auto_configure" mapstructure:"auto_configure"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level string `yaml:"level" mapstructure:"level"`
}

// GetRequiredPorts returns a list of ports that need to be opened
func (c *Config) GetRequiredPorts() []PortInfo {
	var ports []PortInfo

	if c.HTTP.Enabled {
		ports = append(ports, PortInfo{Port: c.HTTP.Port, Protocol: "tcp", Service: "HTTP Proxy"})
	}

	if c.HTTPS.Enabled {
		ports = append(ports, PortInfo{Port: c.HTTPS.Port, Protocol: "tcp", Service: "HTTPS Proxy"})
	}

	if c.Shadowsocks.Enabled {
		ports = append(ports, PortInfo{Port: c.Shadowsocks.Port, Protocol: "tcp", Service: "Shadowsocks"})
		ports = append(ports, PortInfo{Port: c.Shadowsocks.Port, Protocol: "udp", Service: "Shadowsocks"})
	}

	return ports
}

// PortInfo represents information about a network port
type PortInfo struct {
	Port     int
	Protocol string
	Service  string
}
