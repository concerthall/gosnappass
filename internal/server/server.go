package server

import (
	"fmt"
	"io"
	"net/http"
	"os"

	// TODO: mux is deprecated, but until we find something else
	// we'll use it.

	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
)

const (
	defaultListenAddress = ":5000"
)

type Server struct {
	listenAddress  string
	router         *mux.Router
	logHandler     slog.Handler
	logger         *slog.Logger
	pathPrefix     string
	hostOverride   string
	proto          string
	redisKeyPrefix string
}

type ServerOption = func(*Server)

// New returns a server with the provided opts, if any.
func New(opts ...ServerOption) *Server {
	s := Server{
		listenAddress:  defaultListenAddress,
		logHandler:     slog.NewJSONHandler(os.Stdout),
		redisKeyPrefix: "snappass",
	}

	for _, opt := range opts {
		opt(&s)
	}

	// Add the Logger
	s.logger = slog.New(s.logHandler)
	s.router = router(s.logger, routerConfig{
		hostOverride:   s.hostOverride,
		proto:          s.proto,
		pathPrefix:     s.pathPrefix,
		redisKeyPrefix: s.redisKeyPrefix,
	})
	return &s
}

func (srv *Server) Run() error {
	// Check that Redis is running before we start.
	if err := RedisIsAlive(); err != nil {
		return fmt.Errorf("unable to talk to the database: %s", err)
	}

	srv.logger.Info("starting server",
		"listenAddress", srv.listenAddress,
		"pathPrefix", srv.pathPrefix,
		"proto", srv.proto,
		"hostOverride", srv.hostOverride,
	)

	return http.ListenAndServe(
		srv.listenAddress,
		srv.router,
	)
}

// Shutdown executes shutdown logic for the server instance.
func (srv *Server) Shutdown() error {
	return RedisCloseConnections()
}

// WithRedisKeyPrefix instructs the server to prefix all keys inserted into the
// database with prefix.
func WithRedisKeyPrefix(prefix string) ServerOption {
	return func(s *Server) { s.redisKeyPrefix = prefix }
}

// WithPathPrefix informs the server that links should include prefix
// in the URI. This does not change path handling, and only impacts
// link rendering when providing the user with their credential link.
func WithPathPrefix(prefix string) ServerOption {
	return func(s *Server) {
		s.pathPrefix = prefix
	}
}

// WithHostOverride informs the server that links should override the
// default host value with override.
func WithHostOverride(override string) ServerOption {
	return func(s *Server) {
		s.hostOverride = override
	}
}

// WithProto sets the proto returned for password links. Should be http or https
func WithProto(proto string) ServerOption {
	return func(s *Server) {
		s.proto = proto
	}
}

// SetListenAddress sets the server's listen address to address
func SetListenAddress(address string) ServerOption {
	return func(s *Server) {
		s.listenAddress = address
	}
}

// LogTo sets where to log server logs.
func LogTo(w io.Writer) ServerOption {
	return func(s *Server) {
		s.logHandler = slog.NewJSONHandler(w)
	}
}
