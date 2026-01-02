package cli

import (
	"os"

	"github.com/spf13/cobra"

	"wte/internal/config"
	"wte/internal/gost"
	"wte/internal/security"
	"wte/internal/system"
	"wte/internal/ui"
)

var (
	uninstallForce     bool
	uninstallKeepCreds bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall GOST proxy server",
	Long: `Completely remove GOST proxy server and all related files.

This command will:
  - Stop the GOST service
  - Disable autostart
  - Remove the systemd service file
  - Remove the GOST binary
  - Remove configuration files
  - Optionally keep credentials file

Examples:
  wte uninstall              # Uninstall with confirmation
  wte uninstall --force      # Uninstall without confirmation
  wte uninstall --keep-creds # Keep credentials file`,
	RunE: runUninstall,
}

func init() {
	uninstallCmd.Flags().BoolVarP(&uninstallForce, "force", "f", false, "Skip confirmation prompt")
	uninstallCmd.Flags().BoolVar(&uninstallKeepCreds, "keep-creds", false, "Keep credentials file")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	// Check root
	if err := checkRoot(); err != nil {
		return err
	}

	ui.PrintBanner(Version)
	ui.Header("Uninstalling WTE Proxy")

	// Confirmation
	if !uninstallForce {
		ui.Warning("This will completely remove the GOST proxy server installation.")
		ui.Println()
		if !ui.Confirm("Are you sure you want to continue?") {
			ui.Info("Uninstall cancelled")
			return nil
		}
	}

	cfg := config.Get()
	systemd := system.NewSystemdManager()
	osInfo, _ := system.DetectOS()

	var installer *gost.Installer
	if osInfo != nil {
		installer = gost.NewInstaller(cfg, osInfo)
	}

	totalSteps := 6
	currentStep := 0

	// Step 1: Stop service
	currentStep++
	ui.Step(currentStep, totalSteps, "Stopping service")

	status, _ := systemd.Status()
	if status != nil && status.IsActive {
		ui.Action("Stopping GOST service...")
		if err := systemd.Stop(); err != nil {
			ui.Warning("Could not stop service: %v", err)
		} else {
			ui.Success("Service stopped")
		}
	} else {
		ui.Success("Service was not running")
	}

	// Step 2: Disable service
	currentStep++
	ui.Step(currentStep, totalSteps, "Disabling service")

	if status != nil && status.IsEnabled {
		ui.Action("Disabling service autostart...")
		if err := systemd.Disable(); err != nil {
			ui.Warning("Could not disable service: %v", err)
		} else {
			ui.Success("Service disabled")
		}
	} else {
		ui.Success("Service was not enabled")
	}

	// Step 3: Remove systemd service file
	currentStep++
	ui.Step(currentStep, totalSteps, "Removing systemd service")

	if systemd.IsInstalled() {
		ui.Action("Removing service file...")
		if err := systemd.Remove(); err != nil {
			ui.Warning("Could not remove service file: %v", err)
		} else {
			ui.Success("Service file removed")
		}
	} else {
		ui.Success("Service file not found")
	}

	// Step 4: Remove GOST binary
	currentStep++
	ui.Step(currentStep, totalSteps, "Removing GOST binary")

	if installer != nil && installer.IsInstalled() {
		ui.Action("Removing binary...")
		if err := installer.Uninstall(); err != nil {
			ui.Warning("Could not remove binary: %v", err)
		} else {
			ui.Success("Binary removed")
		}
	} else {
		ui.Success("Binary not found")
	}

	// Step 5: Remove configuration
	currentStep++
	ui.Step(currentStep, totalSteps, "Removing configuration")

	configGen := gost.NewConfigGenerator(cfg)
	if err := configGen.Remove(); err != nil {
		ui.Warning("Could not remove GOST configuration: %v", err)
	} else {
		ui.Success("GOST configuration removed")
	}

	// Remove WTE config
	if system.FileExists(config.WTEConfigFile) {
		if err := os.Remove(config.WTEConfigFile); err != nil {
			ui.Warning("Could not remove WTE configuration: %v", err)
		} else {
			ui.Success("WTE configuration removed")
		}
	}

	// Remove TLS certificates if they exist
	if security.CertificateExists(cfg.HTTPS.CertPath, cfg.HTTPS.KeyPath) {
		if err := security.RemoveCertificates(cfg.HTTPS.CertPath, cfg.HTTPS.KeyPath); err != nil {
			ui.Warning("Could not remove certificates: %v", err)
		} else {
			ui.Success("TLS certificates removed")
		}
	}

	// Step 6: Remove credentials file
	currentStep++
	ui.Step(currentStep, totalSteps, "Cleaning up")

	if !uninstallKeepCreds {
		credsMgr := gost.NewCredentialsManager(cfg, "")
		if credsMgr.Exists() {
			if err := credsMgr.Remove(); err != nil {
				ui.Warning("Could not remove credentials file: %v", err)
			} else {
				ui.Success("Credentials file removed")
			}
		}
	} else {
		ui.Info("Keeping credentials file as requested")
	}

	// Done
	ui.Println()
	ui.Green.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	ui.Green.Println("║                    ✓ UNINSTALL COMPLETED SUCCESSFULLY                       ║")
	ui.Green.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	ui.Println()

	ui.Info("GOST proxy server has been completely removed.")
	if uninstallKeepCreds {
		ui.Detail("Credentials file kept at: %s", config.CredentialsFile)
	}

	return nil
}
