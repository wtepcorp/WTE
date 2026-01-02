package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// CertificateOptions holds options for certificate generation
type CertificateOptions struct {
	CommonName   string
	Organization string
	Country      string
	ValidDays    int
	IPAddresses  []string
	DNSNames     []string
	KeyPath      string
	CertPath     string
}

// DefaultCertificateOptions returns default certificate options
func DefaultCertificateOptions(ip string) *CertificateOptions {
	return &CertificateOptions{
		CommonName:   ip,
		Organization: "WTE Proxy",
		Country:      "XX",
		ValidDays:    365,
		IPAddresses:  []string{ip, "127.0.0.1"},
		DNSNames:     []string{"localhost"},
	}
}

// GenerateSelfSignedCert generates a self-signed TLS certificate
func GenerateSelfSignedCert(opts *CertificateOptions) error {
	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Prepare certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(opts.ValidDays) * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   opts.CommonName,
			Organization: []string{opts.Organization},
			Country:      []string{opts.Country},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// Add IP addresses
	for _, ipStr := range opts.IPAddresses {
		if ip := net.ParseIP(ipStr); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		}
	}

	// Add DNS names
	template.DNSNames = opts.DNSNames

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(opts.CertPath), 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Write certificate
	certOut, err := os.Create(opts.CertPath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Set certificate permissions
	if err := os.Chmod(opts.CertPath, 0644); err != nil {
		return fmt.Errorf("failed to set certificate permissions: %w", err)
	}

	// Marshal private key
	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Write private key
	keyOut, err := os.Create(opts.KeyPath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Set key permissions (restricted)
	if err := os.Chmod(opts.KeyPath, 0600); err != nil {
		return fmt.Errorf("failed to set key permissions: %w", err)
	}

	return nil
}

// CertificateExists checks if certificate files exist
func CertificateExists(certPath, keyPath string) bool {
	if _, err := os.Stat(certPath); err != nil {
		return false
	}
	if _, err := os.Stat(keyPath); err != nil {
		return false
	}
	return true
}

// GetCertificateInfo returns information about a certificate
func GetCertificateInfo(certPath string) (*CertificateInfo, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	info := &CertificateInfo{
		Subject:    cert.Subject.CommonName,
		Issuer:     cert.Issuer.CommonName,
		NotBefore:  cert.NotBefore,
		NotAfter:   cert.NotAfter,
		IsExpired:  time.Now().After(cert.NotAfter),
		DaysLeft:   int(time.Until(cert.NotAfter).Hours() / 24),
		IPAddresses: make([]string, 0, len(cert.IPAddresses)),
		DNSNames:   cert.DNSNames,
	}

	for _, ip := range cert.IPAddresses {
		info.IPAddresses = append(info.IPAddresses, ip.String())
	}

	return info, nil
}

// CertificateInfo holds information about a certificate
type CertificateInfo struct {
	Subject     string
	Issuer      string
	NotBefore   time.Time
	NotAfter    time.Time
	IsExpired   bool
	DaysLeft    int
	IPAddresses []string
	DNSNames    []string
}

// RemoveCertificates removes certificate and key files
func RemoveCertificates(certPath, keyPath string) error {
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
