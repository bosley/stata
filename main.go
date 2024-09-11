package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Define command-line flags
	port := flag.Int("port", 8080, "Port to serve on")
	bind := flag.String("bind", "127.0.0.1", "Address to bind to")
	dir := flag.String("directory", ".", "Directory to serve")
	secure := flag.Bool("secure", false, "Enable HTTPS with a self-signed certificate")

	// Parse the flags
	flag.Parse()

	// Get the absolute path of the directory
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}

	// Check if the directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", absDir)
	}

	// Create a file server handler
	fs := http.FileServer(http.Dir(absDir))

	// Set up the handler
	http.Handle("/", fs)

	// Prepare the server
	addr := fmt.Sprintf("%s:%d", *bind, *port)
	var srv *http.Server

	if *secure {
		// Generate temporary certificate and key
		certFile, keyFile, err := generateTempCert()
		if err != nil {
			log.Fatalf("Failed to generate temporary certificate: %v", err)
		}
		defer os.Remove(certFile)
		defer os.Remove(keyFile)

		// Set up HTTPS server
		srv = &http.Server{
			Addr: addr,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}

		// Print startup message
		fmt.Printf("Serving %s on https://%s\n", absDir, addr)

		// Start HTTPS server
		log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
	} else {
		// Print startup message
		fmt.Printf("Serving %s on http://%s\n", absDir, addr)

		// Start HTTP server
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}

func generateTempCert() (string, string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "stata_cert")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Generate a private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %v", err)
	}

	// Prepare certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Stata Temporary Cert"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	// Create the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %v", err)
	}

	// Write the certificate to a temporary file
	certFile := filepath.Join(tempDir, "cert.pem")
	certOut, err := os.Create(certFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to open cert.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", fmt.Errorf("failed to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		return "", "", fmt.Errorf("error closing cert.pem: %v", err)
	}

	// Write the private key to a temporary file
	keyFile := filepath.Join(tempDir, "key.pem")
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", fmt.Errorf("failed to open key.pem for writing: %v", err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return "", "", fmt.Errorf("failed to write data to key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		return "", "", fmt.Errorf("error closing key.pem: %v", err)
	}

	return certFile, keyFile, nil
}
