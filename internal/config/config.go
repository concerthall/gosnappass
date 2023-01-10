package config

import (
	"os"
	"strconv"
)

const (
	// These environment variables intend to align with the original snappass. In the future
	// this project will break binary compatible and provide a migration path.

	EnvRedisURL     = "REDIS_URL"
	EnvRedisHost    = "REDIS_HOST"
	EnvRedisPort    = "REDIS_PORT"
	EnvRedisDB      = "SNAPPASS_REDIS_DB"
	EnvRedisPrefix  = "REDIS_PREFIX"
	EnvStaticURL    = "STATIC_URL"
	EnvURLPrefix    = "URL_PREFIX"
	EnvHostOverride = "HOST_OVERRIDE"
	EnvNoSSL        = "NO_SSL"

	// EnvListenString is the server address on which to listen.,
	// e.g. 192.168.10.10:5000, :1234, etc.
	// The python implementation uses flask which has other environment variables that we
	// will not implement.
	EnvListenString = "SNAPPASS_LISTEN_ADDRESS"
)

func RedisURL() string {
	return os.Getenv(EnvRedisURL)
}

// RedisConnectionOptions returns environment-provided host, port, and db
// values, or defaults if those are unset.
func RedisConnectionOptions() (host, port string, db int) {
	// default values
	host = "localhost"
	port = "6379"
	db = 0

	if h := os.Getenv(EnvRedisHost); h != "" {
		host = h
	}

	if p := os.Getenv(EnvRedisPort); p != "" {
		port = p
	}

	if d := os.Getenv(EnvRedisDB); d != "" {
		if i, err := strconv.Atoi(d); err == nil {
			db = i
		}
	}

	return
}
