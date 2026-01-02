package gost

import (
	"fmt"
	"os"
	"text/template"
	"time"

	"wte/internal/config"
)

const credentialsTemplate = `╔══════════════════════════════════════════════════════════════════════════════╗
║                         PROXY SERVER CREDENTIALS                              ║
╠══════════════════════════════════════════════════════════════════════════════╣
║                                                                               ║
║  Generated: {{.GeneratedAt}}
║  Server IP: {{.ServerIP}}
║  Generator: WTE
║                                                                               ║
╚══════════════════════════════════════════════════════════════════════════════╝
{{if .HTTP.Enabled}}
┌──────────────────────────────────────────────────────────────────────────────┐
│ HTTP PROXY                                                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Host:     {{.ServerIP}}
│  Port:     {{.HTTP.Port}}
{{- if .HTTP.Auth.Enabled}}
│  Username: {{.HTTP.Auth.Username}}
│  Password: {{.HTTP.Auth.Password}}
│                                                                               │
│  Full URL: http://{{.HTTP.Auth.Username}}:{{.HTTP.Auth.Password}}@{{.ServerIP}}:{{.HTTP.Port}}
{{- else}}
│  Authentication: Disabled
│                                                                               │
│  Full URL: http://{{.ServerIP}}:{{.HTTP.Port}}
{{- end}}
│                                                                               │
│  Test command:                                                                │
{{- if .HTTP.Auth.Enabled}}
│  curl -x http://{{.HTTP.Auth.Username}}:{{.HTTP.Auth.Password}}@{{.ServerIP}}:{{.HTTP.Port}} https://ifconfig.me
{{- else}}
│  curl -x http://{{.ServerIP}}:{{.HTTP.Port}} https://ifconfig.me
{{- end}}
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
{{end}}
{{if .HTTPS.Enabled}}
┌──────────────────────────────────────────────────────────────────────────────┐
│ HTTPS PROXY (TLS encrypted)                                                  │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Host:     {{.ServerIP}}
│  Port:     {{.HTTPS.Port}}
{{- if .HTTPS.Auth.Enabled}}
│  Username: {{.HTTPS.Auth.Username}}
│  Password: {{.HTTPS.Auth.Password}}
{{- end}}
│                                                                               │
│  Note: Uses self-signed certificate. Browser may show security warning.      │
│  Certificate: {{.HTTPS.CertPath}}
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
{{end}}
{{if .Shadowsocks.Enabled}}
┌──────────────────────────────────────────────────────────────────────────────┐
│ SHADOWSOCKS                                                                   │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Server:   {{.ServerIP}}
│  Port:     {{.Shadowsocks.Port}}
│  Password: {{.Shadowsocks.Password}}
│  Method:   {{.Shadowsocks.Method}}
│                                                                               │
│  SS URI (for import):                                                         │
│  {{.ShadowsocksURI}}
│                                                                               │
│  Compatible clients:                                                          │
│  - iOS: Shadowrocket, Surge, Quantumult                                       │
│  - Android: Shadowsocks, v2rayNG                                              │
│  - Windows: Shadowsocks-windows, v2rayN                                       │
│  - macOS: ShadowsocksX-NG, Surge                                              │
│  - Linux: shadowsocks-libev, shadowsocks-rust                                 │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
{{end}}
┌──────────────────────────────────────────────────────────────────────────────┐
│ MANAGEMENT COMMANDS                                                           │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Check status:     wte status                                                 │
│  View logs:        wte logs -f                                                │
│  Restart service:  wte restart                                                │
│  Stop service:     wte stop                                                   │
│  Edit config:      wte config edit                                            │
│  Uninstall:        wte uninstall                                              │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘

`

// CredentialsManager manages credentials file
type CredentialsManager struct {
	cfg      *config.Config
	serverIP string
}

// NewCredentialsManager creates a new CredentialsManager
func NewCredentialsManager(cfg *config.Config, serverIP string) *CredentialsManager {
	return &CredentialsManager{
		cfg:      cfg,
		serverIP: serverIP,
	}
}

// Save saves credentials to file
func (m *CredentialsManager) Save() error {
	tmpl, err := template.New("credentials").Parse(credentialsTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse credentials template: %w", err)
	}

	configGen := NewConfigGenerator(m.cfg)

	data := struct {
		GeneratedAt    string
		ServerIP       string
		HTTP           config.HTTPConfig
		HTTPS          config.HTTPSConfig
		Shadowsocks    config.ShadowsocksConfig
		ShadowsocksURI string
	}{
		GeneratedAt:    time.Now().Format("2006-01-02 15:04:05"),
		ServerIP:       m.serverIP,
		HTTP:           m.cfg.HTTP,
		HTTPS:          m.cfg.HTTPS,
		Shadowsocks:    m.cfg.Shadowsocks,
		ShadowsocksURI: configGen.GetShadowsocksURI(m.serverIP),
	}

	// Use same password for HTTPS if not set
	if m.cfg.HTTPS.Enabled && m.cfg.HTTPS.Auth.Password == "" {
		data.HTTPS.Auth = m.cfg.HTTP.Auth
	}

	file, err := os.Create(config.CredentialsFile)
	if err != nil {
		return fmt.Errorf("failed to create credentials file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	// Set restricted permissions
	if err := os.Chmod(config.CredentialsFile, 0600); err != nil {
		return fmt.Errorf("failed to set credentials file permissions: %w", err)
	}

	return nil
}

// Print prints credentials to stdout
func (m *CredentialsManager) Print() error {
	tmpl, err := template.New("credentials").Parse(credentialsTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse credentials template: %w", err)
	}

	configGen := NewConfigGenerator(m.cfg)

	data := struct {
		GeneratedAt    string
		ServerIP       string
		HTTP           config.HTTPConfig
		HTTPS          config.HTTPSConfig
		Shadowsocks    config.ShadowsocksConfig
		ShadowsocksURI string
	}{
		GeneratedAt:    time.Now().Format("2006-01-02 15:04:05"),
		ServerIP:       m.serverIP,
		HTTP:           m.cfg.HTTP,
		HTTPS:          m.cfg.HTTPS,
		Shadowsocks:    m.cfg.Shadowsocks,
		ShadowsocksURI: configGen.GetShadowsocksURI(m.serverIP),
	}

	// Use same password for HTTPS if not set
	if m.cfg.HTTPS.Enabled && m.cfg.HTTPS.Auth.Password == "" {
		data.HTTPS.Auth = m.cfg.HTTP.Auth
	}

	return tmpl.Execute(os.Stdout, data)
}

// Remove removes the credentials file
func (m *CredentialsManager) Remove() error {
	if err := os.Remove(config.CredentialsFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove credentials file: %w", err)
	}
	return nil
}

// Exists checks if credentials file exists
func (m *CredentialsManager) Exists() bool {
	_, err := os.Stat(config.CredentialsFile)
	return err == nil
}

// GetPath returns the credentials file path
func (m *CredentialsManager) GetPath() string {
	return config.CredentialsFile
}
