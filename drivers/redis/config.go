package redisstore

import (
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
}
