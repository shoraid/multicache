package redis

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildTLSConfig(t *testing.T) {
	origSystemCertPool := systemCertPool
	origLoadX509KeyPair := loadX509KeyPair
	defer func() {
		systemCertPool = origSystemCertPool
		loadX509KeyPair = origLoadX509KeyPair
	}()

	tests := []struct {
		name       string
		cfg        *TLSConfig
		mockSystem func() (*x509.CertPool, error)
		mockKey    func(certFile, keyFile string) (tls.Certificate, error)
		wantErr    bool
		errMsg     string
	}{
		{
			name:    "should return nil when cfg is nil",
			cfg:     nil,
			wantErr: false,
		},
		{
			name: "should return error when system cert pool fails",
			cfg:  &TLSConfig{},
			mockSystem: func() (*x509.CertPool, error) {
				return nil, errors.New("mock system pool fail")
			},
			wantErr: true,
			errMsg:  "failed to load system CAs",
		},
		{
			name: "should return error when client cert loading fails",
			cfg:  &TLSConfig{CertFile: "cert.pem", KeyFile: "key.pem"},
			mockSystem: func() (*x509.CertPool, error) {
				return x509.NewCertPool(), nil
			},
			mockKey: func(certFile, keyFile string) (tls.Certificate, error) {
				return tls.Certificate{}, errors.New("mock: keypair fail")
			},
			wantErr: true,
			errMsg:  "failed to load client certificate",
		},
		{
			name: "should return valid tls.Config when all succeed",
			cfg:  &TLSConfig{ServerName: "localhost"},
			mockSystem: func() (*x509.CertPool, error) {
				return x509.NewCertPool(), nil
			},
			mockKey: func(certFile, keyFile string) (tls.Certificate, error) {
				return tls.Certificate{}, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSystem != nil {
				systemCertPool = tt.mockSystem
			}
			if tt.mockKey != nil {
				loadX509KeyPair = tt.mockKey
			}

			cfg, err := buildTLSConfig(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err, "expected error but got nil")
				assert.Nil(t, cfg, "expected tls.Config to be nil")
				assert.Contains(t, err.Error(), tt.errMsg, "expected error message to contain "+tt.errMsg)
			} else {
				assert.NoError(t, err, "expected no error but got one")
				if tt.cfg != nil {
					assert.NotNil(t, cfg, "expected tls.Config to be created")
					assert.Equal(t, tt.cfg.ServerName, cfg.ServerName, "expected ServerName to match")
				}
			}
		})
	}
}

func TestDefaultMinVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    uint16
		expected uint16
	}{
		{
			name:     "should return TLS 1.2 when input is zero",
			input:    0,
			expected: tls.VersionTLS12,
		},
		{
			name:     "should return same value when input is non-zero",
			input:    tls.VersionTLS13,
			expected: tls.VersionTLS13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultMinVersion(tt.input)
			assert.Equal(t, tt.expected, result, "expected result to match")
		})
	}
}
