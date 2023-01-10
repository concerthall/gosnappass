package server

import (
	"context"
	"fmt"

	"github.com/concerthall/gosnappass/internal/config"
	"github.com/go-redis/redis/v8"
)

var clientOpts *redis.Options
var client *redis.Client

// RedisClient returns the redis client based on the environment.
func RedisClient() *redis.Client {
	if client == nil {
		configureRedis()
		client = redis.NewClient(clientOpts)
	}

	return client
}

// RedisCloseConnections closes connections to the server.
func RedisCloseConnections() error {
	return client.Close()
}

// RedisIsAlive returns an error if we cannot communicate with redis.
func RedisIsAlive() error {
	redis := RedisClient()
	pingstatus := redis.Ping(context.TODO())
	return pingstatus.Err()
}

// configureRedis builds out client options based on the application's
// configuration.
func configureRedis() {
	if url := config.RedisURL(); url != "" {
		opt, err := redis.ParseURL(url)
		// if we hit an error, we fallback. No need to return it here.
		if err == nil {
			clientOpts = opt
			return
		}
	}

	host, port, db := config.RedisConnectionOptions()
	clientOpts = &redis.Options{
		DB:   db,
		Addr: fmt.Sprintf("%s:%s", host, port),
	}
}
