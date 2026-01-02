package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"wte/internal/config"
	"wte/internal/ui"
)

// Version information
var (
	Version   = "0.0.1"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var (
	cfgFile   string
	verbose   bool
	quiet     bool
	noColor   bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "wte",
	Short: "WTE - Window to Europe. Manage GOST proxy infrastructure",
	Long: `WTE (Window to Europe) - CLI tool for managing GOST-based proxy infrastructure.

It provides easy commands to install, configure, and manage HTTP, HTTPS,
and Shadowsocks proxy services.

Examples:
  wte install              Install and configure proxy server
  wte status               Show service status
  wte start                Start the proxy service
  wte stop                 Stop the proxy service
  wte logs -f              Follow service logs
  wte config show          Show current configuration
  wte credentials          Show connection credentials`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set UI options
		ui.SetNoColor(noColor)
		ui.SetQuiet(quiet)
		ui.SetVerbose(verbose)

		// Initialize configuration
		if err := config.Init(cfgFile); err != nil {
			// Only warn if config file doesn't exist - it's expected for new installs
			ui.Debug("Config initialization: %v", err)
		}

		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is /etc/wte/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output (only errors)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(credentialsCmd)
}

// checkRoot ensures the command is run as root
func checkRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this command must be run as root")
	}
	return nil
}

// versionCmd shows version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("WTE v%s\n", Version)
		if verbose {
			fmt.Printf("  Build Time: %s\n", BuildTime)
			fmt.Printf("  Git Commit: %s\n", GitCommit)
		}
	},
}
