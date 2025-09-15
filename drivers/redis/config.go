package redis

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"time"
)

// RedisConfig keeps the settings to set up Redis connection.
// This is a simplified version of redis.Options, keeping only
// the most useful fields for production workloads.
type RedisConfig struct {

	// Addr is the address formatted as host:port.
	Addr string

	// ClientName will execute the `CLIENT SETNAME ClientName` command for each conn.
	ClientName string

	// Username is used to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	Username string

	// Password is an optional password. Must match the password specified in the
	// `requirepass` server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string

	// DB is the database to be selected after connecting to the server.
	DB int

	// MaxRetries is the maximum number of retries before giving up.
	// -1 disables retries.
	//
	// default: 3 retries
	MaxRetries int

	// MinRetryBackoff is the minimum backoff between each retry.
	// -1 disables backoff.
	//
	// default: 8 milliseconds
	MinRetryBackoff time.Duration

	// MaxRetryBackoff is the maximum backoff between each retry.
	// -1 disables backoff.
	//
	// default: 512 milliseconds
	MaxRetryBackoff time.Duration

	// DialTimeout for establishing new connections.
	//
	// default: 5 seconds
	DialTimeout time.Duration

	// ReadTimeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking.
	// Supported values:
	//   - `-1` - no timeout (block indefinitely)
	//   - `-2` - disables SetReadDeadline calls completely
	//
	// default: 3 seconds
	ReadTimeout time.Duration

	// WriteTimeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Supported values:
	//   - `-1` - no timeout (block indefinitely)
	//   - `-2` - disables SetWriteDeadline calls completely
	//
	// default: 3 seconds
	WriteTimeout time.Duration

	// PoolSize is the base number of socket connections.
	// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	// If there are not enough connections in the pool, new connections will be allocated
	// in excess of PoolSize.
	//
	// default: 10 * runtime.GOMAXPROCS(0)
	PoolSize int

	// PoolTimeout is the amount of time client waits for a connection if all connections
	// are busy before returning an error.
	//
	// default: ReadTimeout + 1 second
	PoolTimeout time.Duration

	// MinIdleConns is the minimum number of idle connections kept alive in the pool.
	// Useful when establishing new connections is slow.
	//
	// default: 0
	MinIdleConns int

	// ConnMaxIdleTime is the maximum amount of time a connection may be idle.
	// Should be less than the server's timeout.
	//
	// default: 30 minutes
	ConnMaxIdleTime time.Duration

	// ConnMaxLifetime is the maximum amount of time a connection may be reused.
	//
	// default: 0 (no limit)
	ConnMaxLifetime time.Duration

	// UseTLS enables TLS negotiation. When true, TLSConfig will be applied if provided.
	UseTLS bool

	// TLSConfig provides detailed TLS configuration if UseTLS is true.
	// If nil, a default TLS configuration with system root CAs will be used.
	TLSConfig *TLSConfig
}

// TLSConfig is a simplified version of tls.Config for typical client use cases.
// It preserves the order and comments of crypto/tls.Config but keeps only
// the most commonly used fields for Redis and other client connections.
type TLSConfig struct {
	// Rand provides the source of entropy for nonces and RSA blinding.
	// If Rand is nil, TLS uses the cryptographic random reader in package crypto/rand.
	Rand io.Reader

	// Time returns the current time as the number of seconds since the epoch.
	// If Time is nil, TLS uses time.Now.
	Time func() time.Time

	// Certificates contains one or more certificate chains to present to the
	// server (for mutual TLS authentication).
	Certificates []tls.Certificate

	// CAFile is an optional path to a custom CA certificate file (PEM).
	// If provided, it will be loaded and appended to RootCAs.
	CAFile string

	// RootCAs defines the set of root certificate authorities that clients use
	// when verifying server certificates. If RootCAs is nil, system CAs are used.
	RootCAs *x509.CertPool

	// ServerName is used to verify the hostname on the server certificate
	// unless InsecureSkipVerify is true.
	ServerName string

	// InsecureSkipVerify controls whether the client verifies the server's
	// certificate chain and hostname. This should only be used for local
	// development or in combination with a custom verification callback.
	InsecureSkipVerify bool

	// MinVersion contains the minimum TLS version that is acceptable.
	//
	// default: TLS 1.2
	MinVersion uint16

	// MaxVersion contains the maximum TLS version that is acceptable.
	//
	// default: TLS 1.3
	MaxVersion uint16

	// CertFile and KeyFile are optional paths for client certificate (mutual TLS).
	CertFile string
	KeyFile  string
}
