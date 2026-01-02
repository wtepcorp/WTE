package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"wte/internal/ui"
	"wte/internal/updater"
)

var (
	updateCheck bool
	updateForce bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update WTE to the latest version",
	Long: `Check for updates and install the latest version of WTE.

This command will:
  - Check GitHub releases for the latest version
  - Download the appropriate binary for your platform
  - Replace the current binary with the new one

Examples:
  wte update              # Update to latest version
  wte update --check      # Only check for updates
  wte update --force      # Force update even if on latest`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Only check for updates, don't install")
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Force update even if already on latest")

	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	upd := updater.NewUpdater(Version)

	ui.Action("Checking for updates...")

	release, hasUpdate, err := upd.CheckForUpdate()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	currentVersion := Version
	latestVersion := release.TagName

	if !hasUpdate && !updateForce {
		ui.Success("You are already on the latest version (%s)", currentVersion)
		return nil
	}

	if hasUpdate {
		ui.Info("New version available!")
		ui.Println()
		ui.Printf("  Current version: %s\n", currentVersion)
		ui.Printf("  Latest version:  %s\n", latestVersion)
		ui.Println()

		// Show release notes
		notes := upd.GetReleaseNotes(release)
		if notes != "" && notes != "No release notes available." {
			ui.Info("Release notes:")
			ui.Println()
			ui.Printf("%s\n", notes)
			ui.Println()
		}
	} else {
		ui.Info("Forcing reinstall of version %s", latestVersion)
	}

	// If only checking, stop here
	if updateCheck {
		ui.Println()
		ui.Info("Run 'wte update' to install the update")
		return nil
	}

	// Confirm update
	if !updateForce && !ui.Confirm("Do you want to update?") {
		ui.Info("Update cancelled")
		return nil
	}

	// Check root for installation
	if err := checkRoot(); err != nil {
		return fmt.Errorf("update requires root privileges: %w", err)
	}

	ui.Println()
	ui.Header("Updating WTE")

	if err := upd.Update(release); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	ui.Println()
	ui.Green.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	ui.Green.Println("║                        ✓ UPDATE COMPLETED SUCCESSFULLY                      ║")
	ui.Green.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	ui.Println()

	ui.Info("WTE has been updated to version %s", latestVersion)
	ui.Detail("Run 'wte version' to verify")

	return nil
}
