package server

import (
	"net/http"

	"github.com/concerthall/gosnappass/internal/embedded"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
)

// router provides the configured router that's used for handling requests.
func router(logger *slog.Logger, cfg routerConfig) *mux.Router {
	m := mux.NewRouter()

	// Handle static files. Note that we do not StripPrefix /static/ from the filesystem
	// directory because it matches what the HTTP path is.
	fs := http.FileServer(http.FS(embedded.StaticAssets))
	m.PathPrefix("/static/").Handler(fs)

	// Handle the Favicon separately because we want to return a 404 if, for whatever reason
	// the embedding fails.
	m.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// if it's empty, we want to 404
		if len(embedded.Favicon) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, _ = w.Write(embedded.Favicon)
	})

	// Register all other handlers.
	m.HandleFunc("/{token}", showConfirmationHandler).Methods(http.MethodGet)
	m.HandleFunc("/{token}", getPasswordHandler).Methods(http.MethodPost)
	m.HandleFunc("/", indexHandler).Methods(http.MethodGet)
	m.HandleFunc("/", newSetPasswordHandler(cfg.proto, cfg.hostOverride, cfg.redisKeyPrefix, cfg.pathPrefix)).Methods(http.MethodPost)

	m.Use(
		addRequestIDMW,
		newInjectLoggerMW(logger),
		logRequestMW,
		newDatabasePingMW(RedisIsAlive),
	)

	return m
}

type routerConfig struct {
	hostOverride   string
	proto          string
	pathPrefix     string
	redisKeyPrefix string
}
