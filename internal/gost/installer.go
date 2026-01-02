package gost

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"wte/internal/config"
	"wte/internal/system"
	"wte/internal/ui"
)

const (
	// GOSTGitHubURL is the base URL for GOST releases
	GOSTGitHubURL = "https://github.com/go-gost/gost/releases/download"
)

// Installer handles GOST installation
type Installer struct {
	cfg    *config.Config
	osInfo *system.OSInfo
}

// NewInstaller creates a new Installer
func NewInstaller(cfg *config.Config, osInfo *system.OSInfo) *Installer {
	return &Installer{
		cfg:    cfg,
		osInfo: osInfo,
	}
}

// Install downloads and installs GOST
func (i *Installer) Install() error {
	version := i.cfg.GOST.Version
	arch := i.osInfo.GOSTArch

	ui.Action("Downloading GOST v%s for %s...", version, arch)

	// Construct download URL
	downloadURL := fmt.Sprintf("%s/v%s/gost_%s_linux_%s.tar.gz",
		GOSTGitHubURL, version, version, arch)

	ui.Detail("URL: %s", downloadURL)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "gost_install_")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, "gost.tar.gz")

	// Download archive
	if err := i.downloadFile(archivePath, downloadURL); err != nil {
		return fmt.Errorf("failed to download GOST: %w", err)
	}

	ui.Success("Download completed")

	// Extract archive
	ui.Action("Extracting archive...")
	if err := i.extractTarGz(archivePath, tempDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	ui.Success("Archive extracted")

	// Find gost binary in extracted files
	gostBinary := filepath.Join(tempDir, "gost")
	if !system.FileExists(gostBinary) {
		return fmt.Errorf("gost binary not found in archive")
	}

	// Install binary
	ui.Action("Installing GOST binary to %s...", i.cfg.GOST.BinaryPath)

	// Ensure target directory exists
	targetDir := filepath.Dir(i.cfg.GOST.BinaryPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create binary directory: %w", err)
	}

	// Copy binary
	if err := i.copyFile(gostBinary, i.cfg.GOST.BinaryPath); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(i.cfg.GOST.BinaryPath, 0755); err != nil {
		return fmt.Errorf("failed to set binary permissions: %w", err)
	}

	ui.Success("GOST binary installed")

	// Verify installation
	ui.Action("Verifying installation...")
	version, err = i.GetVersion()
	if err != nil {
		return fmt.Errorf("failed to verify installation: %w", err)
	}

	ui.Success("GOST installed successfully")
	ui.Detail("Version: %s", version)

	return nil
}

// downloadFile downloads a file with progress
func (i *Installer) downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create progress bar
	bar := ui.DownloadProgressBar(resp.ContentLength, "gost.tar.gz")
	defer bar.Finish()

	// Copy with progress
	_, err = io.Copy(io.MultiWriter(out, bar.Writer()), resp.Body)
	return err
}

// extractTarGz extracts a tar.gz archive
func (i *Installer) extractTarGz(archive, dest string) error {
	file, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func (i *Installer) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// GetVersion returns the installed GOST version
func (i *Installer) GetVersion() (string, error) {
	if !i.IsInstalled() {
		return "", fmt.Errorf("GOST is not installed")
	}

	cmd := exec.Command(i.cfg.GOST.BinaryPath, "-V")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// IsInstalled checks if GOST is installed
func (i *Installer) IsInstalled() bool {
	return system.FileExists(i.cfg.GOST.BinaryPath)
}

// Uninstall removes GOST binary
func (i *Installer) Uninstall() error {
	if !i.IsInstalled() {
		return nil
	}

	if err := os.Remove(i.cfg.GOST.BinaryPath); err != nil {
		return fmt.Errorf("failed to remove GOST binary: %w", err)
	}

	return nil
}

// GetLatestVersion fetches the latest GOST version from GitHub
func (i *Installer) GetLatestVersion() (string, error) {
	// This would require GitHub API call
	// For now, return the configured version
	return i.cfg.GOST.Version, nil
}

// NeedsUpdate checks if GOST needs to be updated
func (i *Installer) NeedsUpdate() (bool, string, error) {
	if !i.IsInstalled() {
		return true, i.cfg.GOST.Version, nil
	}

	currentVersion, err := i.GetVersion()
	if err != nil {
		return false, "", err
	}

	latestVersion, err := i.GetLatestVersion()
	if err != nil {
		return false, "", err
	}

	return currentVersion != latestVersion, latestVersion, nil
}
