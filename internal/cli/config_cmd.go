package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"wte/internal/config"
	"wte/internal/gost"
	"wte/internal/system"
	"wte/internal/ui"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `View and modify WTE proxy configuration.

Subcommands:
  show     Show current configuration
  edit     Open configuration in editor
  set      Set a configuration value
  reset    Reset configuration to defaults

Examples:
  wte config show
  wte config edit
  wte config set http.port 3128
  wte config set http.auth.enabled false`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()

		ui.Header("Current Configuration")

		// Display as YAML
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Println(string(data))

		ui.Println()
		ui.Detail("Config file: %s", config.GetConfigPath())

		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration in editor",
	Long: `Open the WTE configuration file in your default editor.

The editor is determined by:
1. $EDITOR environment variable
2. $VISUAL environment variable
3. Fallback to 'nano' or 'vi'

After saving, you should restart the service:
  wte restart`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkRoot(); err != nil {
			return err
		}

		configPath := config.WTEConfigFile

		// Ensure config file exists
		if !system.FileExists(configPath) {
			ui.Action("Creating default configuration...")
			cfg := config.DefaultConfig()
			if err := config.SaveTo(configPath); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
			_ = cfg // silence unused warning
		}

		// Find editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			// Try common editors
			for _, e := range []string{"nano", "vi", "vim"} {
				if _, err := exec.LookPath(e); err == nil {
					editor = e
					break
				}
			}
		}

		if editor == "" {
			return fmt.Errorf("no editor found. Set $EDITOR environment variable")
		}

		ui.Info("Opening %s with %s...", configPath, editor)

		editCmd := exec.Command(editor, configPath)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr

		if err := editCmd.Run(); err != nil {
			return fmt.Errorf("editor exited with error: %w", err)
		}

		ui.Println()
		ui.Success("Configuration saved")
		ui.Info("Run 'wte restart' to apply changes")

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  http.enabled          Enable/disable HTTP proxy (true/false)
  http.port             HTTP proxy port
  http.auth.enabled     Enable/disable HTTP authentication (true/false)
  http.auth.username    HTTP proxy username
  http.auth.password    HTTP proxy password

  https.enabled         Enable/disable HTTPS proxy (true/false)
  https.port            HTTPS proxy port

  shadowsocks.enabled   Enable/disable Shadowsocks (true/false)
  shadowsocks.port      Shadowsocks port
  shadowsocks.method    Shadowsocks encryption method
  shadowsocks.password  Shadowsocks password

  firewall.auto_configure  Auto-configure firewall (true/false)

Examples:
  wte config set http.port 3128
  wte config set http.auth.enabled false
  wte config set shadowsocks.enabled true`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkRoot(); err != nil {
			return err
		}

		key := args[0]
		value := args[1]

		// Parse value based on key
		var parsedValue interface{}
		switch {
		case strings.HasSuffix(key, ".enabled"):
			parsedValue = value == "true" || value == "1" || value == "yes"
		case strings.HasSuffix(key, ".port"):
			var port int
			if _, err := fmt.Sscanf(value, "%d", &port); err != nil {
				return fmt.Errorf("invalid port number: %s", value)
			}
			parsedValue = port
		default:
			parsedValue = value
		}

		if err := config.Set(key, parsedValue); err != nil {
			return fmt.Errorf("failed to set configuration: %w", err)
		}

		if err := config.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		ui.Success("Configuration updated: %s = %v", key, parsedValue)
		ui.Info("Run 'wte restart' to apply changes")

		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long: `Reset all configuration to default values.

This will:
1. Generate new random passwords
2. Reset all ports to defaults
3. Reset all options to defaults

Examples:
  wte config reset`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkRoot(); err != nil {
			return err
		}

		if !ui.Confirm("Reset configuration to defaults?") {
			ui.Info("Reset cancelled")
			return nil
		}

		config.Reset()

		if err := config.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		ui.Success("Configuration reset to defaults")
		ui.Info("Run 'wte restart' to apply changes")

		return nil
	},
}

var configApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply configuration changes",
	Long: `Regenerate GOST configuration from WTE config and restart service.

This command:
1. Reads current WTE configuration
2. Regenerates GOST config.yaml
3. Restarts the GOST service

Examples:
  wte config apply`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkRoot(); err != nil {
			return err
		}

		cfg := config.Get()

		ui.Action("Regenerating GOST configuration...")

		configGen := gost.NewConfigGenerator(cfg)
		if err := configGen.Generate(); err != nil {
			return fmt.Errorf("failed to generate configuration: %w", err)
		}

		ui.Success("Configuration regenerated")

		ui.Action("Restarting service...")
		systemd := system.NewSystemdManager()
		if err := systemd.Restart(); err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}

		ui.Success("Service restarted")

		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configApplyCmd)
}
