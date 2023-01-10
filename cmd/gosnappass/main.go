package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/concerthall/gosnappass/internal/config"
	"github.com/concerthall/gosnappass/internal/server"
)

func main() {
	serverOptions := []server.ServerOption{}

	if val, isSet := os.LookupEnv(config.EnvHostOverride); isSet {
		serverOptions = append(serverOptions, server.WithHostOverride(val))
	}

	if val, isSet := os.LookupEnv(config.EnvURLPrefix); isSet {
		serverOptions = append(serverOptions, server.WithPathPrefix(val))
	}

	if val, isSet := os.LookupEnv(config.EnvNoSSL); isSet && strings.ToLower(val) == "false" {
		serverOptions = append(serverOptions, server.WithProto("https"))
	}

	if val, isSet := os.LookupEnv(config.EnvListenString); isSet {
		serverOptions = append(serverOptions, server.SetListenAddress(val))
	}

	if val, isSet := os.LookupEnv(config.EnvRedisPrefix); isSet {
		serverOptions = append(serverOptions, server.WithRedisKeyPrefix(val))
	}

	srv := server.New(serverOptions...)

	// handle OS signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// run the application
	go func() {
		if err := srv.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "The server quit with error: "+err.Error())
			os.Exit(1)
		}
	}()

	s := <-signals
	fmt.Println("received signal:", s, "-- shutting down")
	if err := srv.Shutdown(); err != nil {
		fmt.Println(err)
		os.Exit(9)
	}
}
