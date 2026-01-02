package system

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"wte/internal/config"
)

const systemdServiceTemplate = `# ============================================================================
# GOST Proxy Server - Systemd Service Unit
# ============================================================================
# Managed by WTE
# Do not edit manually - changes may be overwritten
# ============================================================================

[Unit]
Description=GOST Proxy Server (WTE)
Documentation=https://gost.run/
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart={{.BinaryPath}} -C {{.ConfigFile}}
Restart=always
RestartSec=5
LimitNOFILE=65535

# Security Hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths={{.ConfigDir}}
PrivateTmp=true
PrivateDevices=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

[Install]
WantedBy=multi-user.target
`

// ServiceStatus represents the status of a systemd service
type ServiceStatus struct {
	Name        string
	IsActive    bool
	IsEnabled   bool
	MainPID     string
	MemoryUsage string
	ActiveState string
	SubState    string
	LoadState   string
}

// SystemdManager manages systemd services
type SystemdManager struct{}

// NewSystemdManager creates a new SystemdManager
func NewSystemdManager() *SystemdManager {
	return &SystemdManager{}
}

// CreateService creates the systemd service file
func (m *SystemdManager) CreateService(cfg *config.Config) error {
	tmpl, err := template.New("service").Parse(systemdServiceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	data := struct {
		BinaryPath string
		ConfigFile string
		ConfigDir  string
	}{
		BinaryPath: cfg.GOST.BinaryPath,
		ConfigFile: cfg.GOST.ConfigFile,
		ConfigDir:  cfg.GOST.ConfigDir,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute service template: %w", err)
	}

	if err := os.WriteFile(config.SystemdServiceFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	return nil
}

// DaemonReload reloads the systemd daemon
func (m *SystemdManager) DaemonReload() error {
	return m.runSystemctl("daemon-reload")
}

// Enable enables the service for autostart
func (m *SystemdManager) Enable() error {
	return m.runSystemctl("enable", "gost")
}

// Disable disables the service autostart
func (m *SystemdManager) Disable() error {
	return m.runSystemctl("disable", "gost")
}

// Start starts the service
func (m *SystemdManager) Start() error {
	return m.runSystemctl("start", "gost")
}

// Stop stops the service
func (m *SystemdManager) Stop() error {
	return m.runSystemctl("stop", "gost")
}

// Restart restarts the service
func (m *SystemdManager) Restart() error {
	return m.runSystemctl("restart", "gost")
}

// Reload reloads the service configuration
func (m *SystemdManager) Reload() error {
	return m.runSystemctl("reload", "gost")
}

// Status returns the service status
func (m *SystemdManager) Status() (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name: "gost",
	}

	// Check if active
	if err := m.runSystemctl("is-active", "--quiet", "gost"); err == nil {
		status.IsActive = true
	}

	// Check if enabled
	if err := m.runSystemctl("is-enabled", "--quiet", "gost"); err == nil {
		status.IsEnabled = true
	}

	// Get detailed status
	output, err := m.getSystemctlOutput("show", "gost",
		"--property=ActiveState,SubState,LoadState,MainPID,MemoryCurrent")
	if err == nil {
		for _, line := range strings.Split(output, "\n") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			switch parts[0] {
			case "ActiveState":
				status.ActiveState = parts[1]
			case "SubState":
				status.SubState = parts[1]
			case "LoadState":
				status.LoadState = parts[1]
			case "MainPID":
				status.MainPID = parts[1]
			case "MemoryCurrent":
				if parts[1] != "[not set]" {
					// Convert bytes to MB
					var bytes int64
					_, _ = fmt.Sscanf(parts[1], "%d", &bytes)
					status.MemoryUsage = fmt.Sprintf("%dMB", bytes/1024/1024)
				}
			}
		}
	}

	return status, nil
}

// IsInstalled checks if the service is installed
func (m *SystemdManager) IsInstalled() bool {
	return FileExists(config.SystemdServiceFile)
}

// Remove removes the service file
func (m *SystemdManager) Remove() error {
	if !m.IsInstalled() {
		return nil
	}

	// Stop and disable first (ignore errors as service might not be running)
	_ = m.Stop()
	_ = m.Disable()

	// Remove service file
	if err := os.Remove(config.SystemdServiceFile); err != nil {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload daemon
	return m.DaemonReload()
}

// GetLogs returns recent service logs
func (m *SystemdManager) GetLogs(lines int) (string, error) {
	args := []string{"-u", "gost", "-n", fmt.Sprintf("%d", lines), "--no-pager"}
	return m.getJournalctlOutput(args...)
}

// FollowLogs starts following logs and returns a command that can be waited on
func (m *SystemdManager) FollowLogs() *exec.Cmd {
	cmd := exec.Command("journalctl", "-u", "gost", "-f", "--no-pager")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// runSystemctl runs a systemctl command
func (m *SystemdManager) runSystemctl(args ...string) error {
	cmd := exec.Command("systemctl", args...)
	return cmd.Run()
}

// getSystemctlOutput runs a systemctl command and returns output
func (m *SystemdManager) getSystemctlOutput(args ...string) (string, error) {
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getJournalctlOutput runs a journalctl command and returns output
func (m *SystemdManager) getJournalctlOutput(args ...string) (string, error) {
	cmd := exec.Command("journalctl", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// IsSystemd checks if the system uses systemd
func IsSystemd() bool {
	_, err := os.Stat("/run/systemd/system")
	return err == nil
}
