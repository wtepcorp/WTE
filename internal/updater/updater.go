package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"wte/internal/ui"
)

const (
	// GitHubRepo is the repository path
	GitHubRepo = "wtepcorp/WTE"

	// GitHubAPIURL is the GitHub API base URL
	GitHubAPIURL = "https://api.github.com"

	// ReleasesURL is the URL for releases
	ReleasesURL = GitHubAPIURL + "/repos/" + GitHubRepo + "/releases"
)

// Release represents a GitHub release
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
	HTMLURL     string    `json:"html_url"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}

// Updater handles self-update functionality
type Updater struct {
	currentVersion string
	repoURL        string
	httpClient     *http.Client
}

// NewUpdater creates a new Updater
func NewUpdater(currentVersion string) *Updater {
	return &Updater{
		currentVersion: currentVersion,
		repoURL:        GitHubRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetRepoURL sets a custom repository URL
func (u *Updater) SetRepoURL(repo string) {
	u.repoURL = repo
}

// GetLatestRelease fetches the latest release from GitHub
func (u *Updater) GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", GitHubAPIURL, u.repoURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "wte-updater")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	return &release, nil
}

// CheckForUpdate checks if an update is available
func (u *Updater) CheckForUpdate() (*Release, bool, error) {
	release, err := u.GetLatestRelease()
	if err != nil {
		return nil, false, err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(u.currentVersion, "v")

	// Simple version comparison (works for semver)
	hasUpdate := latestVersion != currentVersion && latestVersion > currentVersion

	return release, hasUpdate, nil
}

// GetAssetForPlatform finds the appropriate asset for the current platform
func (u *Updater) GetAssetForPlatform(release *Release) (*Asset, error) {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Expected asset name pattern: wte-linux-amd64.tar.gz
	expectedName := fmt.Sprintf("wte-%s-%s.tar.gz", os, arch)
	expectedNameAlt := fmt.Sprintf("wte_%s_%s.tar.gz", os, arch)

	for _, asset := range release.Assets {
		if asset.Name == expectedName || asset.Name == expectedNameAlt {
			return &asset, nil
		}
	}

	// Try without extension for direct binary
	expectedBinary := fmt.Sprintf("wte-%s-%s", os, arch)
	expectedBinaryAlt := fmt.Sprintf("wte_%s_%s", os, arch)

	for _, asset := range release.Assets {
		if asset.Name == expectedBinary || asset.Name == expectedBinaryAlt {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("no asset found for %s/%s", os, arch)
}

// DownloadAsset downloads a release asset
func (u *Updater) DownloadAsset(asset *Asset, destPath string) error {
	resp, err := u.httpClient.Get(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	// Create progress bar
	bar := ui.DownloadProgressBar(asset.Size, asset.Name)

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(io.MultiWriter(out, bar.Writer()), resp.Body)
	bar.Finish()

	return err
}

// Update performs the self-update
func (u *Updater) Update(release *Release) error {
	asset, err := u.GetAssetForPlatform(release)
	if err != nil {
		return err
	}

	ui.Action("Downloading %s...", asset.Name)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "wte-update-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	downloadPath := filepath.Join(tempDir, asset.Name)

	// Download asset
	if err := u.DownloadAsset(asset, downloadPath); err != nil {
		return err
	}

	ui.Success("Download completed")

	// Extract if it's a tarball
	var binaryPath string
	if strings.HasSuffix(asset.Name, ".tar.gz") || strings.HasSuffix(asset.Name, ".tgz") {
		ui.Action("Extracting archive...")
		binaryPath, err = u.extractTarGz(downloadPath, tempDir)
		if err != nil {
			return fmt.Errorf("failed to extract: %w", err)
		}
		ui.Success("Extracted")
	} else {
		binaryPath = downloadPath
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	ui.Action("Installing new version...")

	// Backup current binary
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Copy new binary
	if err := u.copyFile(binaryPath, execPath); err != nil {
		// Restore backup on failure
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(execPath, 0755); err != nil {
		// Restore backup on failure
		_ = os.Remove(execPath)
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Remove backup
	_ = os.Remove(backupPath)

	ui.Success("Updated to version %s", release.TagName)

	return nil
}

// extractTarGz extracts a tar.gz archive and returns the path to the binary
func (u *Updater) extractTarGz(archive, dest string) (string, error) {
	file, err := os.Open(archive)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var binaryPath string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			// Look for the wte binary
			if header.Name == "wte" || strings.HasSuffix(header.Name, "/wte") {
				binaryPath = target
			}

			outFile, err := os.Create(target)
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", err
			}
			outFile.Close()

			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return "", err
			}
		}
	}

	if binaryPath == "" {
		// Look for any executable
		entries, _ := os.ReadDir(dest)
		for _, entry := range entries {
			if !entry.IsDir() && (entry.Name() == "wte" || strings.HasPrefix(entry.Name(), "wte")) {
				binaryPath = filepath.Join(dest, entry.Name())
				break
			}
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("binary not found in archive")
	}

	return binaryPath, nil
}

// copyFile copies a file from src to dst
func (u *Updater) copyFile(src, dst string) error {
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

// GetReleaseNotes returns formatted release notes
func (u *Updater) GetReleaseNotes(release *Release) string {
	if release.Body == "" {
		return "No release notes available."
	}
	return release.Body
}
