package cli

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"wte/internal/system"
	"wte/internal/ui"
)

var (
	logsFollow bool
	logsLines  int
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View service logs",
	Long: `View GOST proxy service logs from journald.

Examples:
  wte logs              # Show last 50 lines
  wte logs -n 100       # Show last 100 lines
  wte logs -f           # Follow logs in real-time
  wte logs -f -n 20     # Follow with 20 initial lines`,
	RunE: runLogs,
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().IntVarP(&logsLines, "lines", "n", 50, "Number of lines to show")
}

func runLogs(cmd *cobra.Command, args []string) error {
	systemd := system.NewSystemdManager()

	if !systemd.IsInstalled() {
		return fmt.Errorf("service is not installed")
	}

	if logsFollow {
		// Follow logs
		ui.Info("Following logs... (press Ctrl+C to stop)")
		ui.Println()

		logCmd := systemd.FollowLogs()
		if err := logCmd.Start(); err != nil {
			return fmt.Errorf("failed to start log stream: %w", err)
		}

		// Handle interrupt signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			_ = logCmd.Process.Kill()
		}()

		if err := logCmd.Wait(); err != nil {
			// Ignore signal-related errors
			if exitErr, ok := err.(*exec.ExitError); ok {
				// Process was killed by signal (user pressed Ctrl+C)
				if !exitErr.Success() {
					return nil
				}
			}
		}
	} else {
		// Show recent logs
		logs, err := systemd.GetLogs(logsLines)
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}

		if logs == "" {
			ui.Info("No logs available")
			return nil
		}

		fmt.Print(logs)
	}

	return nil
}
