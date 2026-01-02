package system

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// IPServices is a list of services to query for public IP
var IPServices = []string{
	"https://ifconfig.me",
	"https://icanhazip.com",
	"https://ipinfo.io/ip",
	"https://api.ipify.org",
	"https://ipecho.net/plain",
}

// GetPublicIP attempts to determine the public IP address
func GetPublicIP() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)

	for _, service := range IPServices {
		resp, err := client.Get(service)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		ip := strings.TrimSpace(string(body))
		if ipRegex.MatchString(ip) {
			return ip, nil
		}
	}

	return "", fmt.Errorf("could not determine public IP address")
}

// GetLocalIPs returns a list of local IP addresses
func GetLocalIPs() ([]string, error) {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

// IsPortOpen checks if a port is listening
func IsPortOpen(port int) bool {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// IsPortAvailable checks if a port is available for binding
func IsPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// GetListeningPorts returns a map of ports to process names
func GetListeningPorts() map[int]string {
	// This is a simplified version - in production you'd parse /proc/net/tcp
	// or use ss/netstat output
	return make(map[int]string)
}

// CheckConnectivity verifies internet connectivity
func CheckConnectivity() bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://www.google.com")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

// ResolveHostname resolves a hostname to IP addresses
func ResolveHostname(hostname string) ([]string, error) {
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return nil, err
	}
	return addrs, nil
}

// GetDefaultGateway attempts to get the default gateway
func GetDefaultGateway() (string, error) {
	// This would require parsing /proc/net/route or using netlink
	// Simplified version that might not work on all systems
	return "", fmt.Errorf("not implemented")
}
