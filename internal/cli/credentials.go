package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"wte/internal/config"
	"wte/internal/gost"
	"wte/internal/security"
	"wte/internal/system"
	"wte/internal/ui"
)

var (
	credsRegenerate bool
	credsShowURI    bool
)

var credentialsCmd = &cobra.Command{
	Use:     "credentials",
	Aliases: []string{"creds"},
	Short:   "Show or regenerate proxy credentials",
	Long: `Display or regenerate proxy server credentials.

This command shows:
  - HTTP proxy connection details
  - HTTPS proxy connection details (if enabled)
  - Shadowsocks connection details (if enabled)
  - Shadowsocks URI for mobile clients

Examples:
  wte credentials              # Show credentials
  wte creds                    # Short alias
  wte credentials --regenerate # Generate new passwords
  wte credentials --uri        # Show Shadowsocks URI only`,
	RunE: runCredentials,
}

func init() {
	credentialsCmd.Flags().BoolVarP(&credsRegenerate, "regenerate", "r", false, "Regenerate passwords")
	credentialsCmd.Flags().BoolVar(&credsShowURI, "uri", false, "Show Shadowsocks URI only")
}

func runCredentials(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	// Get public IP
	publicIP, err := system.GetPublicIP()
	if err != nil {
		ui.Warning("Could not detect public IP: %v", err)
		publicIP = "YOUR_SERVER_IP"
	}

	// Regenerate passwords if requested
	if credsRegenerate {
		if err := checkRoot(); err != nil {
			return err
		}

		ui.Action("Regenerating passwords...")

		// Generate new HTTP password
		if cfg.HTTP.Auth.Enabled {
			pass, err := security.GeneratePassword(16)
			if err != nil {
				return fmt.Errorf("failed to generate HTTP password: %w", err)
			}
			cfg.HTTP.Auth.Password = pass
			cfg.HTTPS.Auth.Password = pass
		}

		// Generate new Shadowsocks password
		if cfg.Shadowsocks.Enabled {
			pass, err := security.GeneratePassword(16)
			if err != nil {
				return fmt.Errorf("failed to generate Shadowsocks password: %w", err)
			}
			cfg.Shadowsocks.Password = pass
		}

		// Save configuration
		if err := config.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		// Regenerate GOST config
		configGen := gost.NewConfigGenerator(cfg)
		if err := configGen.Generate(); err != nil {
			return fmt.Errorf("failed to regenerate GOST config: %w", err)
		}

		// Save credentials file
		credsMgr := gost.NewCredentialsManager(cfg, publicIP)
		if err := credsMgr.Save(); err != nil {
			ui.Warning("Could not save credentials file: %v", err)
		}

		// Restart service
		ui.Action("Restarting service...")
		systemd := system.NewSystemdManager()
		if err := systemd.Restart(); err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}

		ui.Success("Passwords regenerated and service restarted")
		ui.Println()
	}

	// Show Shadowsocks URI only
	if credsShowURI {
		if !cfg.Shadowsocks.Enabled {
			return fmt.Errorf("Shadowsocks is not enabled")
		}

		configGen := gost.NewConfigGenerator(cfg)
		uri := configGen.GetShadowsocksURI(publicIP)
		fmt.Println(uri)
		return nil
	}

	// Print full credentials
	credsMgr := gost.NewCredentialsManager(cfg, publicIP)
	return credsMgr.Print()
}
