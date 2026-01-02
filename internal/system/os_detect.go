package system

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// OSInfo contains information about the operating system
type OSInfo struct {
	OS           string // ubuntu, debian, centos, etc.
	Version      string // 22.04, 11, 8, etc.
	PrettyName   string // Ubuntu 22.04.3 LTS
	Arch         string // x86_64, aarch64, armv7l
	GOSTArch     string // amd64, arm64, armv7
	IsSupported  bool
	PackageManager string // apt, yum, dnf, pacman
}

// DetectOS detects the operating system and architecture
func DetectOS() (*OSInfo, error) {
	info := &OSInfo{
		Arch: runtime.GOARCH,
	}

	// Detect architecture
	if err := detectArch(info); err != nil {
		return nil, err
	}

	// Detect OS
	if err := detectOSRelease(info); err != nil {
		// Try fallback for older systems
		if err := detectRedHatRelease(info); err != nil {
			return nil, fmt.Errorf("unable to detect operating system: %w", err)
		}
	}

	// Set package manager
	setPackageManager(info)

	// Check if supported
	checkSupported(info)

	return info, nil
}

// detectArch detects the system architecture
func detectArch(info *OSInfo) error {
	// Use uname -m equivalent
	arch := runtime.GOARCH

	switch arch {
	case "amd64":
		info.Arch = "x86_64"
		info.GOSTArch = "amd64"
	case "arm64":
		info.Arch = "aarch64"
		info.GOSTArch = "arm64"
	case "arm":
		info.Arch = "armv7l"
		info.GOSTArch = "armv7"
	default:
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	return nil
}

// detectOSRelease reads /etc/os-release
func detectOSRelease(info *OSInfo) error {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := strings.Trim(parts[1], `"`)

		switch key {
		case "ID":
			info.OS = value
		case "VERSION_ID":
			info.Version = value
		case "PRETTY_NAME":
			info.PrettyName = value
		}
	}

	if info.OS == "" {
		return fmt.Errorf("could not determine OS ID")
	}

	return scanner.Err()
}

// detectRedHatRelease is a fallback for older RHEL-based systems
func detectRedHatRelease(info *OSInfo) error {
	data, err := os.ReadFile("/etc/redhat-release")
	if err != nil {
		return err
	}

	content := string(data)
	info.OS = "centos"
	info.PrettyName = strings.TrimSpace(content)

	// Try to extract version
	parts := strings.Fields(content)
	for _, part := range parts {
		if strings.Contains(part, ".") {
			info.Version = strings.Split(part, ".")[0]
			break
		}
	}

	return nil
}

// setPackageManager determines the package manager based on OS
func setPackageManager(info *OSInfo) {
	switch info.OS {
	case "ubuntu", "debian", "linuxmint", "pop":
		info.PackageManager = "apt"
	case "centos", "rhel", "rocky", "almalinux", "oracle":
		if info.Version >= "8" {
			info.PackageManager = "dnf"
		} else {
			info.PackageManager = "yum"
		}
	case "fedora":
		info.PackageManager = "dnf"
	case "arch", "manjaro", "endeavouros":
		info.PackageManager = "pacman"
	case "opensuse", "opensuse-leap", "opensuse-tumbleweed":
		info.PackageManager = "zypper"
	case "alpine":
		info.PackageManager = "apk"
	default:
		info.PackageManager = "unknown"
	}
}

// checkSupported marks whether the OS is officially supported
func checkSupported(info *OSInfo) {
	supported := map[string]bool{
		"ubuntu":    true,
		"debian":    true,
		"centos":    true,
		"rhel":      true,
		"rocky":     true,
		"almalinux": true,
		"fedora":    true,
		"arch":      true,
		"manjaro":   true,
	}

	info.IsSupported = supported[info.OS]
}

// IsRoot checks if the current user is root
func IsRoot() bool {
	return os.Geteuid() == 0
}

// GetHostname returns the system hostname
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string, perm os.FileMode) error {
	if !DirExists(path) {
		return os.MkdirAll(path, perm)
	}
	return nil
}
