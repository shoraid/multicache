package redis

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Test seams so we can mock in unit tests
var systemCertPool = x509.SystemCertPool
var loadX509KeyPair = tls.LoadX509KeyPair

var buildTLSConfig = func(cfg *TLSConfig) (*tls.Config, error) {
	if cfg == nil {
		return nil, nil
	}

	// Use system CAs if none provided
	rootCAs := cfg.RootCAs
	if rootCAs == nil {
		pool, err := systemCertPool() // <--- use seam
		if err != nil {
			return nil, fmt.Errorf("failed to load system CAs: %w", err)
		}
		rootCAs = pool
	}

	// Optionally load custom CA file
	if cfg.CAFile != "" {
		caCert, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}
		if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("no valid CA certificates found in %s", cfg.CAFile)
		}
	}

	// Optionally load client cert for mTLS
	var certs []tls.Certificate
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := loadX509KeyPair(cfg.CertFile, cfg.KeyFile) // <--- use seam
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		certs = append(certs, cert)
	}

	return &tls.Config{
		Rand:               cfg.Rand,
		Time:               cfg.Time,
		Certificates:       certs,
		RootCAs:            rootCAs,
		ServerName:         cfg.ServerName,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		MinVersion:         defaultMinVersion(cfg.MinVersion),
		MaxVersion:         cfg.MaxVersion,
	}, nil
}

func defaultMinVersion(v uint16) uint16 {
	if v == 0 {
		return tls.VersionTLS12
	}

	return v
}
