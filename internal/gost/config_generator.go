package gost

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"wte/internal/config"
	"wte/internal/ui"
)

const gostConfigTemplate = `# ============================================================================
# GOST Proxy Server Configuration
# ============================================================================
# Generated: {{.GeneratedAt}}
# Generator: WTE
# Documentation: https://gost.run/
# ============================================================================

services:
{{- if .HTTP.Enabled}}

  # --------------------------------------------------------------------------
  # HTTP Proxy Service
  # --------------------------------------------------------------------------
  {{- if .HTTP.Auth.Enabled}}
  # Authentication: ENABLED
  # URL: http://{{.HTTP.Auth.Username}}:{{.HTTP.Auth.Password}}@SERVER:{{.HTTP.Port}}
  {{- else}}
  # Authentication: DISABLED
  # URL: http://SERVER:{{.HTTP.Port}}
  {{- end}}
  # --------------------------------------------------------------------------
  - name: http-proxy
    addr: ":{{.HTTP.Port}}"
    handler:
      type: http
      {{- if .HTTP.Auth.Enabled}}
      auth:
        username: {{.HTTP.Auth.Username}}
        password: {{.HTTP.Auth.Password}}
      {{- end}}
    listener:
      type: tcp
{{- end}}

{{- if .HTTPS.Enabled}}

  # --------------------------------------------------------------------------
  # HTTPS Proxy Service (TLS encrypted)
  # --------------------------------------------------------------------------
  # Certificate: {{.HTTPS.CertPath}}
  # Key: {{.HTTPS.KeyPath}}
  # --------------------------------------------------------------------------
  - name: https-proxy
    addr: ":{{.HTTPS.Port}}"
    handler:
      type: http
      {{- if .HTTPS.Auth.Enabled}}
      auth:
        username: {{.HTTPS.Auth.Username}}
        password: {{.HTTPS.Auth.Password}}
      {{- end}}
    listener:
      type: tls
      tls:
        certFile: {{.HTTPS.CertPath}}
        keyFile: {{.HTTPS.KeyPath}}
{{- end}}

{{- if .Shadowsocks.Enabled}}

  # --------------------------------------------------------------------------
  # Shadowsocks Service
  # --------------------------------------------------------------------------
  # Server: SERVER:{{.Shadowsocks.Port}}
  # Password: {{.Shadowsocks.Password}}
  # Method: {{.Shadowsocks.Method}}
  # --------------------------------------------------------------------------
  - name: shadowsocks
    addr: ":{{.Shadowsocks.Port}}"
    handler:
      type: ss
      auth:
        username: {{.Shadowsocks.Method}}
        password: {{.Shadowsocks.Password}}
    listener:
      type: tcp
{{- end}}
`

// ConfigGenerator generates GOST configuration
type ConfigGenerator struct {
	cfg *config.Config
}

// NewConfigGenerator creates a new ConfigGenerator
func NewConfigGenerator(cfg *config.Config) *ConfigGenerator {
	return &ConfigGenerator{cfg: cfg}
}

// Generate generates the GOST configuration file
func (g *ConfigGenerator) Generate() error {
	ui.Action("Generating GOST configuration...")

	// Ensure config directory exists
	configDir := filepath.Dir(g.cfg.GOST.ConfigFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Parse template
	tmpl, err := template.New("gost-config").Parse(gostConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}

	// Prepare template data
	data := struct {
		GeneratedAt string
		HTTP        config.HTTPConfig
		HTTPS       config.HTTPSConfig
		Shadowsocks config.ShadowsocksConfig
	}{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		HTTP:        g.cfg.HTTP,
		HTTPS:       g.cfg.HTTPS,
		Shadowsocks: g.cfg.Shadowsocks,
	}

	// If HTTPS uses same auth as HTTP, copy it
	if g.cfg.HTTPS.Enabled && g.cfg.HTTPS.Auth.Password == "" {
		data.HTTPS.Auth = g.cfg.HTTP.Auth
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute config template: %w", err)
	}

	// Write configuration file
	if err := os.WriteFile(g.cfg.GOST.ConfigFile, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	ui.Success("Configuration file created: %s", g.cfg.GOST.ConfigFile)

	// Log summary
	g.logConfigSummary()

	return nil
}

// logConfigSummary logs a summary of the configuration
func (g *ConfigGenerator) logConfigSummary() {
	ui.Info("Configuration summary:")

	if g.cfg.HTTP.Enabled {
		authStatus := "disabled"
		if g.cfg.HTTP.Auth.Enabled {
			authStatus = fmt.Sprintf("user=%s", g.cfg.HTTP.Auth.Username)
		}
		ui.Detail("HTTP Proxy: :%d (%s)", g.cfg.HTTP.Port, authStatus)
	}

	if g.cfg.HTTPS.Enabled {
		ui.Detail("HTTPS Proxy: :%d", g.cfg.HTTPS.Port)
	}

	if g.cfg.Shadowsocks.Enabled {
		ui.Detail("Shadowsocks: :%d (method=%s)", g.cfg.Shadowsocks.Port, g.cfg.Shadowsocks.Method)
	}
}

// Validate validates the configuration
func (g *ConfigGenerator) Validate() error {
	if !g.cfg.HTTP.Enabled && !g.cfg.HTTPS.Enabled && !g.cfg.Shadowsocks.Enabled {
		return fmt.Errorf("at least one service must be enabled")
	}

	// Check port conflicts
	ports := make(map[int]string)

	if g.cfg.HTTP.Enabled {
		if existing, ok := ports[g.cfg.HTTP.Port]; ok {
			return fmt.Errorf("port %d conflict: HTTP and %s", g.cfg.HTTP.Port, existing)
		}
		ports[g.cfg.HTTP.Port] = "HTTP"
	}

	if g.cfg.HTTPS.Enabled {
		if existing, ok := ports[g.cfg.HTTPS.Port]; ok {
			return fmt.Errorf("port %d conflict: HTTPS and %s", g.cfg.HTTPS.Port, existing)
		}
		ports[g.cfg.HTTPS.Port] = "HTTPS"
	}

	if g.cfg.Shadowsocks.Enabled {
		if existing, ok := ports[g.cfg.Shadowsocks.Port]; ok {
			return fmt.Errorf("port %d conflict: Shadowsocks and %s", g.cfg.Shadowsocks.Port, existing)
		}
		ports[g.cfg.Shadowsocks.Port] = "Shadowsocks"
	}

	return nil
}

// GetShadowsocksURI generates a Shadowsocks URI for client import
func (g *ConfigGenerator) GetShadowsocksURI(serverIP string) string {
	if !g.cfg.Shadowsocks.Enabled {
		return ""
	}

	// Format: ss://method:password@server:port
	auth := fmt.Sprintf("%s:%s", g.cfg.Shadowsocks.Method, g.cfg.Shadowsocks.Password)
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))

	return fmt.Sprintf("ss://%s@%s:%d#WTE-Proxy",
		encoded, serverIP, g.cfg.Shadowsocks.Port)
}

// Remove removes the GOST configuration file
func (g *ConfigGenerator) Remove() error {
	if err := os.Remove(g.cfg.GOST.ConfigFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	// Remove config directory if empty
	configDir := filepath.Dir(g.cfg.GOST.ConfigFile)
	entries, err := os.ReadDir(configDir)
	if err == nil && len(entries) == 0 {
		os.Remove(configDir)
	}

	return nil
}

// Backup creates a backup of the current configuration
func (g *ConfigGenerator) Backup() (string, error) {
	if _, err := os.Stat(g.cfg.GOST.ConfigFile); os.IsNotExist(err) {
		return "", nil
	}

	backupPath := fmt.Sprintf("%s.backup.%s",
		g.cfg.GOST.ConfigFile,
		time.Now().Format("20060102_150405"))

	data, err := os.ReadFile(g.cfg.GOST.ConfigFile)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}
