package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"wte/internal/config"
	"wte/internal/system"
	"wte/internal/ui"
)

// startCmd starts the proxy service
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the proxy service",
	Long: `Start the GOST proxy service.

Examples:
  wte start`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkRoot(); err != nil {
			return err
		}

		systemd := system.NewSystemdManager()

		if !systemd.IsInstalled() {
			return fmt.Errorf("service is not installed. Run 'wte install' first")
		}

		status, err := systemd.Status()
		if err == nil && status.IsActive {
			ui.Info("Service is already running")
			return nil
		}

		ui.Action("Starting service...")
		if err := systemd.Start(); err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}

		ui.Success("Service started")

		// Show status
		status, err = systemd.Status()
		if err == nil {
			ui.Detail("PID: %s", status.MainPID)
		}

		return nil
	},
}

// stopCmd stops the proxy service
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the proxy service",
	Long: `Stop the GOST proxy service.

Examples:
  wte stop`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkRoot(); err != nil {
			return err
		}

		systemd := system.NewSystemdManager()

		if !systemd.IsInstalled() {
			return fmt.Errorf("service is not installed")
		}

		status, err := systemd.Status()
		if err == nil && !status.IsActive {
			ui.Info("Service is not running")
			return nil
		}

		ui.Action("Stopping service...")
		if err := systemd.Stop(); err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}

		ui.Success("Service stopped")
		return nil
	},
}

// restartCmd restarts the proxy service
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the proxy service",
	Long: `Restart the GOST proxy service.

Examples:
  wte restart`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkRoot(); err != nil {
			return err
		}

		systemd := system.NewSystemdManager()

		if !systemd.IsInstalled() {
			return fmt.Errorf("service is not installed. Run 'wte install' first")
		}

		ui.Action("Restarting service...")
		if err := systemd.Restart(); err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}

		ui.Success("Service restarted")

		// Show status
		status, err := systemd.Status()
		if err == nil {
			ui.Detail("PID: %s", status.MainPID)
		}

		return nil
	},
}

// statusCmd shows service status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show service status",
	Long: `Show the current status of the GOST proxy service.

This command displays:
  - Service status (running/stopped)
  - Process information
  - Listening ports
  - Configuration summary

Examples:
  wte status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		systemd := system.NewSystemdManager()
		cfg := config.Get()

		ui.Header("WTE Proxy Status")

		// Service status
		if !systemd.IsInstalled() {
			ui.Warning("Service is not installed")
			ui.Detail("Run 'wte install' to set up the proxy server")
			return nil
		}

		status, err := systemd.Status()
		if err != nil {
			ui.Warning("Could not get service status: %v", err)
		} else {
			// Status indicator
			if status.IsActive {
				ui.Success("Service: RUNNING")
			} else {
				ui.Error("Service: STOPPED")
			}

			ui.Detail("State: %s (%s)", status.ActiveState, status.SubState)
			ui.Detail("Enabled: %v", status.IsEnabled)

			if status.MainPID != "" && status.MainPID != "0" {
				ui.Detail("PID: %s", status.MainPID)
			}

			if status.MemoryUsage != "" {
				ui.Detail("Memory: %s", status.MemoryUsage)
			}
		}

		ui.Println()

		// Port status
		ui.Info("Listening Ports:")

		ports := cfg.GetRequiredPorts()
		for _, port := range ports {
			if system.IsPortOpen(port.Port) {
				ui.Success("  %s: :%d (%s) - LISTENING", port.Service, port.Port, port.Protocol)
			} else {
				ui.Error("  %s: :%d (%s) - NOT LISTENING", port.Service, port.Port, port.Protocol)
			}
		}

		ui.Println()

		// Configuration summary
		ui.Info("Configuration:")
		ui.Detail("Config file: %s", config.GetConfigPath())

		if cfg.HTTP.Enabled {
			authStatus := "disabled"
			if cfg.HTTP.Auth.Enabled {
				authStatus = fmt.Sprintf("user=%s", cfg.HTTP.Auth.Username)
			}
			ui.Detail("HTTP Proxy: :%d (%s)", cfg.HTTP.Port, authStatus)
		}

		if cfg.HTTPS.Enabled {
			ui.Detail("HTTPS Proxy: :%d", cfg.HTTPS.Port)
		}

		if cfg.Shadowsocks.Enabled {
			ui.Detail("Shadowsocks: :%d (method=%s)", cfg.Shadowsocks.Port, cfg.Shadowsocks.Method)
		}

		return nil
	},
}
