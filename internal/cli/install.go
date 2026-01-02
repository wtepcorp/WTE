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
	installHTTPPort       int
	installHTTPUser       string
	installHTTPPass       string
	installHTTPNoAuth     bool
	installSSEnabled      bool
	installSSPort         int
	installSSPassword     string
	installSSMethod       string
	installHTTPSEnabled   bool
	installHTTPSPort      int
	installGOSTVersion    string
	installSkipFirewall   bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install and configure GOST proxy server",
	Long: `Install and configure GOST proxy server with HTTP, HTTPS, and Shadowsocks support.

Examples:
  # Basic installation with defaults
  wte install

  # Custom HTTP proxy port and user
  wte install --http-port 3128 --http-user admin

  # Disable HTTP authentication
  wte install --http-no-auth

  # HTTP proxy only (no Shadowsocks)
  wte install --ss-enabled=false

  # Enable HTTPS proxy
  wte install --https-enabled`,
	RunE: runInstall,
}

func init() {
	// HTTP flags
	installCmd.Flags().IntVar(&installHTTPPort, "http-port", config.DefaultHTTPPort, "HTTP proxy port")
	installCmd.Flags().StringVar(&installHTTPUser, "http-user", config.DefaultUsername, "HTTP proxy username")
	installCmd.Flags().StringVar(&installHTTPPass, "http-pass", "", "HTTP proxy password (auto-generated if empty)")
	installCmd.Flags().BoolVar(&installHTTPNoAuth, "http-no-auth", false, "Disable HTTP proxy authentication")

	// Shadowsocks flags
	installCmd.Flags().BoolVar(&installSSEnabled, "ss-enabled", true, "Enable Shadowsocks")
	installCmd.Flags().IntVar(&installSSPort, "ss-port", config.DefaultShadowsocksPort, "Shadowsocks port")
	installCmd.Flags().StringVar(&installSSPassword, "ss-password", "", "Shadowsocks password (auto-generated if empty)")
	installCmd.Flags().StringVar(&installSSMethod, "ss-method", config.DefaultShadowsocksMethod, "Shadowsocks encryption method")

	// HTTPS flags
	installCmd.Flags().BoolVar(&installHTTPSEnabled, "https-enabled", false, "Enable HTTPS proxy")
	installCmd.Flags().IntVar(&installHTTPSPort, "https-port", config.DefaultHTTPSPort, "HTTPS proxy port")

	// Other flags
	installCmd.Flags().StringVar(&installGOSTVersion, "gost-version", config.DefaultGOSTVersion, "GOST version to install")
	installCmd.Flags().BoolVar(&installSkipFirewall, "skip-firewall", false, "Skip firewall configuration")
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Check root
	if err := checkRoot(); err != nil {
		return err
	}

	// Print banner
	ui.PrintBanner(Version)

	// Total steps
	totalSteps := 9
	currentStep := 0

	// Step 1: Detect OS
	currentStep++
	ui.Step(currentStep, totalSteps, "Detecting operating system")

	osInfo, err := system.DetectOS()
	if err != nil {
		return fmt.Errorf("failed to detect OS: %w", err)
	}

	ui.Success("Operating system detected")
	ui.Detail("Distribution: %s", osInfo.OS)
	ui.Detail("Version: %s", osInfo.Version)
	ui.Detail("Architecture: %s (%s)", osInfo.Arch, osInfo.GOSTArch)

	if !osInfo.IsSupported {
		ui.Warning("OS '%s' is not officially tested", osInfo.OS)
	}

	// Step 2: Get public IP
	currentStep++
	ui.Step(currentStep, totalSteps, "Detecting public IP address")

	publicIP, err := system.GetPublicIP()
	if err != nil {
		ui.Warning("Could not detect public IP: %v", err)
		publicIP = "YOUR_SERVER_IP"
	} else {
		ui.Success("Public IP detected: %s", publicIP)
	}

	// Step 3: Prepare configuration
	currentStep++
	ui.Step(currentStep, totalSteps, "Preparing configuration")

	cfg := config.DefaultConfig()

	// Apply command-line options
	cfg.GOST.Version = installGOSTVersion
	cfg.HTTP.Port = installHTTPPort
	cfg.HTTP.Auth.Username = installHTTPUser
	cfg.HTTP.Auth.Enabled = !installHTTPNoAuth

	cfg.Shadowsocks.Enabled = installSSEnabled
	cfg.Shadowsocks.Port = installSSPort
	cfg.Shadowsocks.Method = installSSMethod

	cfg.HTTPS.Enabled = installHTTPSEnabled
	cfg.HTTPS.Port = installHTTPSPort

	cfg.Firewall.AutoConfigure = !installSkipFirewall

	// Generate passwords if needed
	if cfg.HTTP.Auth.Enabled {
		if installHTTPPass != "" {
			cfg.HTTP.Auth.Password = installHTTPPass
		} else {
			pass, err := security.GeneratePassword(16)
			if err != nil {
				return fmt.Errorf("failed to generate HTTP password: %w", err)
			}
			cfg.HTTP.Auth.Password = pass
		}
	}

	if cfg.Shadowsocks.Enabled {
		if installSSPassword != "" {
			cfg.Shadowsocks.Password = installSSPassword
		} else {
			pass, err := security.GeneratePassword(16)
			if err != nil {
				return fmt.Errorf("failed to generate Shadowsocks password: %w", err)
			}
			cfg.Shadowsocks.Password = pass
		}
	}

	// Use same password for HTTPS
	cfg.HTTPS.Auth = cfg.HTTP.Auth

	ui.Success("Configuration prepared")
	ui.Detail("HTTP Proxy: :%d (auth: %v)", cfg.HTTP.Port, cfg.HTTP.Auth.Enabled)
	if cfg.Shadowsocks.Enabled {
		ui.Detail("Shadowsocks: :%d", cfg.Shadowsocks.Port)
	}
	if cfg.HTTPS.Enabled {
		ui.Detail("HTTPS Proxy: :%d", cfg.HTTPS.Port)
	}

	// Step 4: Check existing installation
	currentStep++
	ui.Step(currentStep, totalSteps, "Checking existing installation")

	systemd := system.NewSystemdManager()
	installer := gost.NewInstaller(cfg, osInfo)

	if installer.IsInstalled() {
		ui.Warning("Existing GOST installation detected")

		// Stop service if running
		status, _ := systemd.Status()
		if status != nil && status.IsActive {
			ui.Action("Stopping existing service...")
			if err := systemd.Stop(); err != nil {
				ui.Warning("Could not stop service: %v", err)
			} else {
				ui.Success("Service stopped")
			}
		}

		// Backup config
		configGen := gost.NewConfigGenerator(cfg)
		backupPath, err := configGen.Backup()
		if err != nil {
			ui.Warning("Could not backup configuration: %v", err)
		} else if backupPath != "" {
			ui.Success("Configuration backed up: %s", backupPath)
		}
	} else {
		ui.Success("No existing installation found")
	}

	// Step 5: Install GOST
	currentStep++
	ui.Step(currentStep, totalSteps, "Installing GOST")

	if err := installer.Install(); err != nil {
		return fmt.Errorf("failed to install GOST: %w", err)
	}

	// Step 6: Generate TLS certificates (if HTTPS enabled)
	currentStep++
	ui.Step(currentStep, totalSteps, "Generating TLS certificates")

	if cfg.HTTPS.Enabled {
		ui.Action("Generating self-signed certificate...")

		certOpts := security.DefaultCertificateOptions(publicIP)
		certOpts.CertPath = cfg.HTTPS.CertPath
		certOpts.KeyPath = cfg.HTTPS.KeyPath

		if err := security.GenerateSelfSignedCert(certOpts); err != nil {
			return fmt.Errorf("failed to generate certificate: %w", err)
		}

		ui.Success("TLS certificate generated")
		ui.Detail("Certificate: %s", cfg.HTTPS.CertPath)
		ui.Detail("Private key: %s", cfg.HTTPS.KeyPath)
	} else {
		ui.Success("HTTPS disabled, skipping certificate generation")
	}

	// Step 7: Generate GOST configuration
	currentStep++
	ui.Step(currentStep, totalSteps, "Generating GOST configuration")

	configGen := gost.NewConfigGenerator(cfg)

	if err := configGen.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	if err := configGen.Generate(); err != nil {
		return fmt.Errorf("failed to generate configuration: %w", err)
	}

	// Save WTE configuration
	if err := config.SaveTo(config.WTEConfigFile); err != nil {
		ui.Warning("Could not save WTE configuration: %v", err)
	}

	// Step 8: Create and start systemd service
	currentStep++
	ui.Step(currentStep, totalSteps, "Creating systemd service")

	if err := systemd.CreateService(cfg); err != nil {
		return fmt.Errorf("failed to create systemd service: %w", err)
	}

	ui.Success("Systemd service created")

	ui.Action("Reloading systemd daemon...")
	if err := systemd.DaemonReload(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	ui.Action("Enabling service for autostart...")
	if err := systemd.Enable(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	ui.Action("Starting service...")
	if err := systemd.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	ui.Success("Service started")

	// Verify service status
	status, err := systemd.Status()
	if err != nil {
		ui.Warning("Could not get service status: %v", err)
	} else if status.IsActive {
		ui.Detail("PID: %s", status.MainPID)
		if status.MemoryUsage != "" {
			ui.Detail("Memory: %s", status.MemoryUsage)
		}
	}

	// Step 9: Configure firewall
	currentStep++
	ui.Step(currentStep, totalSteps, "Configuring firewall")

	if cfg.Firewall.AutoConfigure {
		firewall := system.NewFirewallManager()

		ui.Action("Detected firewall: %s", firewall.GetType())

		if err := firewall.OpenPorts(cfg); err != nil {
			ui.Warning("Failed to configure firewall: %v", err)
			ui.Detail("Please manually open required ports")
		} else {
			ui.Success("Firewall configured")
			for _, port := range cfg.GetRequiredPorts() {
				ui.Detail("Port %d/%s opened", port.Port, port.Protocol)
			}
		}
	} else {
		ui.Success("Firewall configuration skipped")
	}

	// Save credentials
	credsMgr := gost.NewCredentialsManager(cfg, publicIP)
	if err := credsMgr.Save(); err != nil {
		ui.Warning("Could not save credentials file: %v", err)
	} else {
		ui.Success("Credentials saved to: %s", credsMgr.GetPath())
	}

	// Print summary
	printInstallSummary(cfg, publicIP)

	return nil
}

func printInstallSummary(cfg *config.Config, publicIP string) {
	ui.Println()
	ui.Green.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	ui.Green.Println("║                    ✓ INSTALLATION COMPLETED SUCCESSFULLY                    ║")
	ui.Green.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	ui.Println()

	// HTTP Proxy
	if cfg.HTTP.Enabled {
		ui.PrintCredentialsBox("HTTP PROXY", map[string]string{
			"Host":     publicIP,
			"Port":     fmt.Sprintf("%d", cfg.HTTP.Port),
			"Username": cfg.HTTP.Auth.Username,
			"Password": cfg.HTTP.Auth.Password,
		})
	}

	// Shadowsocks
	if cfg.Shadowsocks.Enabled {
		ui.PrintCredentialsBox("SHADOWSOCKS", map[string]string{
			"Server":   publicIP,
			"Port":     fmt.Sprintf("%d", cfg.Shadowsocks.Port),
			"Password": cfg.Shadowsocks.Password,
			"Method":   cfg.Shadowsocks.Method,
		})
	}

	ui.Println()
	ui.White.Println("Quick Commands:")
	if cfg.HTTP.Auth.Enabled {
		ui.Printf("  Test:    curl -x http://%s:%s@%s:%d https://ifconfig.me\n",
			cfg.HTTP.Auth.Username, cfg.HTTP.Auth.Password, publicIP, cfg.HTTP.Port)
	} else {
		ui.Printf("  Test:    curl -x http://%s:%d https://ifconfig.me\n",
			publicIP, cfg.HTTP.Port)
	}
	ui.Printf("  Status:  wte status\n")
	ui.Printf("  Logs:    wte logs -f\n")
	ui.Println()
}
