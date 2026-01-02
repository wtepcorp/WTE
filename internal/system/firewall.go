package system

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"wte/internal/config"
)

// FirewallType represents the type of firewall
type FirewallType string

const (
	FirewallUFW       FirewallType = "ufw"
	FirewallFirewalld FirewallType = "firewalld"
	FirewallIPTables  FirewallType = "iptables"
	FirewallNone      FirewallType = "none"
)

// FirewallManager manages firewall rules
type FirewallManager struct {
	firewallType FirewallType
}

// NewFirewallManager creates a new FirewallManager
func NewFirewallManager() *FirewallManager {
	fm := &FirewallManager{}
	fm.detectFirewall()
	return fm
}

// detectFirewall detects which firewall is in use
func (fm *FirewallManager) detectFirewall() {
	// Check for UFW (Ubuntu/Debian)
	if fm.commandExists("ufw") {
		fm.firewallType = FirewallUFW
		return
	}

	// Check for firewalld (CentOS/Fedora/RHEL)
	if fm.commandExists("firewall-cmd") && fm.isServiceActive("firewalld") {
		fm.firewallType = FirewallFirewalld
		return
	}

	// Check for iptables (fallback)
	if fm.commandExists("iptables") {
		fm.firewallType = FirewallIPTables
		return
	}

	fm.firewallType = FirewallNone
}

// GetType returns the detected firewall type
func (fm *FirewallManager) GetType() FirewallType {
	return fm.firewallType
}

// OpenPorts opens the required ports for the proxy
func (fm *FirewallManager) OpenPorts(cfg *config.Config) error {
	ports := cfg.GetRequiredPorts()

	for _, port := range ports {
		if err := fm.OpenPort(port.Port, port.Protocol); err != nil {
			return fmt.Errorf("failed to open port %d/%s: %w", port.Port, port.Protocol, err)
		}
	}

	return fm.Apply()
}

// OpenPort opens a single port
func (fm *FirewallManager) OpenPort(port int, protocol string) error {
	switch fm.firewallType {
	case FirewallUFW:
		return fm.openPortUFW(port, protocol)
	case FirewallFirewalld:
		return fm.openPortFirewalld(port, protocol)
	case FirewallIPTables:
		return fm.openPortIPTables(port, protocol)
	case FirewallNone:
		return nil
	}
	return nil
}

// ClosePort closes a single port
func (fm *FirewallManager) ClosePort(port int, protocol string) error {
	switch fm.firewallType {
	case FirewallUFW:
		return fm.closePortUFW(port, protocol)
	case FirewallFirewalld:
		return fm.closePortFirewalld(port, protocol)
	case FirewallIPTables:
		return fm.closePortIPTables(port, protocol)
	case FirewallNone:
		return nil
	}
	return nil
}

// Apply applies firewall changes (reload)
func (fm *FirewallManager) Apply() error {
	switch fm.firewallType {
	case FirewallUFW:
		// UFW applies changes immediately
		return nil
	case FirewallFirewalld:
		return fm.runCommand("firewall-cmd", "--reload")
	case FirewallIPTables:
		// Try to save rules
		return fm.saveIPTables()
	case FirewallNone:
		return nil
	}
	return nil
}

// Status returns firewall status information
func (fm *FirewallManager) Status() (string, error) {
	switch fm.firewallType {
	case FirewallUFW:
		return fm.getCommandOutput("ufw", "status", "verbose")
	case FirewallFirewalld:
		return fm.getCommandOutput("firewall-cmd", "--list-all")
	case FirewallIPTables:
		return fm.getCommandOutput("iptables", "-L", "-n")
	case FirewallNone:
		return "No firewall detected", nil
	}
	return "", nil
}

// UFW methods
func (fm *FirewallManager) openPortUFW(port int, protocol string) error {
	return fm.runCommand("ufw", "allow", fmt.Sprintf("%d/%s", port, protocol))
}

func (fm *FirewallManager) closePortUFW(port int, protocol string) error {
	return fm.runCommand("ufw", "delete", "allow", fmt.Sprintf("%d/%s", port, protocol))
}

// Firewalld methods
func (fm *FirewallManager) openPortFirewalld(port int, protocol string) error {
	return fm.runCommand("firewall-cmd", "--permanent", "--add-port", fmt.Sprintf("%d/%s", port, protocol))
}

func (fm *FirewallManager) closePortFirewalld(port int, protocol string) error {
	return fm.runCommand("firewall-cmd", "--permanent", "--remove-port", fmt.Sprintf("%d/%s", port, protocol))
}

// IPTables methods
func (fm *FirewallManager) openPortIPTables(port int, protocol string) error {
	return fm.runCommand("iptables", "-A", "INPUT", "-p", protocol, "--dport", strconv.Itoa(port), "-j", "ACCEPT")
}

func (fm *FirewallManager) closePortIPTables(port int, protocol string) error {
	return fm.runCommand("iptables", "-D", "INPUT", "-p", protocol, "--dport", strconv.Itoa(port), "-j", "ACCEPT")
}

func (fm *FirewallManager) saveIPTables() error {
	// Try different methods to persist iptables rules
	if fm.commandExists("netfilter-persistent") {
		return fm.runCommand("netfilter-persistent", "save")
	}
	if FileExists("/etc/iptables") {
		output, err := fm.getCommandOutput("iptables-save")
		if err != nil {
			return err
		}
		return writeFile("/etc/iptables/rules.v4", []byte(output), 0644)
	}
	return nil
}

// Helper methods
func (fm *FirewallManager) commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func (fm *FirewallManager) isServiceActive(name string) bool {
	cmd := exec.Command("systemctl", "is-active", "--quiet", name)
	return cmd.Run() == nil
}

func (fm *FirewallManager) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

func (fm *FirewallManager) getCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func writeFile(path string, data []byte, perm uint32) error {
	return exec.Command("sh", "-c", fmt.Sprintf("echo '%s' > %s", string(data), path)).Run()
}

// Enable enables the firewall
func (fm *FirewallManager) Enable() error {
	switch fm.firewallType {
	case FirewallUFW:
		return fm.runCommand("ufw", "--force", "enable")
	case FirewallFirewalld:
		return fm.runCommand("systemctl", "enable", "--now", "firewalld")
	default:
		return nil
	}
}

// IsEnabled checks if the firewall is enabled
func (fm *FirewallManager) IsEnabled() bool {
	switch fm.firewallType {
	case FirewallUFW:
		output, _ := fm.getCommandOutput("ufw", "status")
		return strings.Contains(output, "Status: active")
	case FirewallFirewalld:
		return fm.isServiceActive("firewalld")
	default:
		return false
	}
}
