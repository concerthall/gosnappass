package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/concerthall/gosnappass/internal/view"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
)

// timeConversion provides text to value translation relative to the UI, allowing
// us to properly include TTL values when adding keys to the database.
var timeConversion = map[string]int{
	"two weeks": 1209600, "week": 604800, "day": 86400, "hour": 3600,
}

// tokenSeparator separates the token ID and the key used to decrypt it in the URL.
const tokenSeparator = "~"

// indexHandler handles requests to /. A user will see a form requesting the credential
// they wish to have stored, and for what duration. GET
func indexHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.FromContext(r.Context())
	if err := view.Index(w); err != nil {
		logger.Error("unable to render index view", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// newSetPasswordHandler produces a setPasswordHandler with proto and hostOverrides to be
// used for the returned link, and keyPrefix used for redis keys.
func newSetPasswordHandler(proto, hostOverride, keyPrefix, urlPrefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := slog.FromContext(r.Context())
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("unable to parse form", err)
			return
		}

		secret := r.FormValue("password")
		ttl := r.FormValue("ttl")

		var ittl int
		var exists bool
		if ittl, exists = timeConversion[strings.ToLower(ttl)]; !exists {
			view.CredentialExpiredOrNotFound(w)
			return
		}

		token, key, err := Encrypt(secret)
		if err != nil {
			logger.Error("error encrypting secret", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id := keyPrefix + uuid.New().String()

		db := RedisClient()
		set := db.Set(r.Context(), id, token, time.Duration(ittl)*time.Second)
		if set.Err() != nil {
			logger.Error("unable to set key with ttl", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// use the host override if set
		host := r.Host
		if hostOverride != "" {
			host = hostOverride
		}
		// normalize the proto TODO: should we do this elsewhere?
		if proto == "" {
			proto = "http"
		}

		link, _ := url.JoinPath(
			fmt.Sprintf("%s://%s/", proto, host),
			urlPrefix,
			strings.Join([]string{id, url.PathEscape(key)}, tokenSeparator))

		if err := view.Confirm(w, link); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// showConfirmationHandler is the UI shown to the user when accessing the access string
// on this server with an http GET request.
func showConfirmationHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.FromContext(r.Context())
	// get the variables from the request's PATH, using gorilla mux's variable system
	vars := mux.Vars(r)

	// split the token into the id and its key
	id, _, err := splitToken(vars["token"])
	if err != nil {
		logger.Error("unable to split token in URL", err)
		view.CredentialExpiredOrNotFound(w)
		return
	}

	db := RedisClient()
	c := db.Exists(r.Context(), id)
	if c.Err() != nil {
		logger.Error("unable to query key from datadbase", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Zero when it doesn't exist in the db
	if c.Val() == 0 {
		view.CredentialExpiredOrNotFound(w)
		return
	}

	if err := view.PreviewPassword(w); err != nil {
		logger.Error("unable to render view PreviewPassword", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// getPasswordHandler is the UI shown to the user containing their password. POST.
func getPasswordHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.FromContext(r.Context())
	// get the variables from the request's PATH, using gorilla mux's variable system
	vars := mux.Vars(r)

	// split the token into the id and its key
	id, key, err := splitToken(vars["token"])
	if err != nil {
		logger.Error("unable to split token in URL", err)
		view.CredentialExpiredOrNotFound(w)
		return
	}

	key, _ = url.PathUnescape(key)

	db := RedisClient()
	c := db.GetDel(r.Context(), id)
	err = c.Err()
	if err != nil {
		// redis.Nil implies the key didn't exist at access time. We'll throw a 404 for this
		// because it's possible the key timed out while another view for this key was loaded.
		if err == redis.Nil {
			view.CredentialExpiredOrNotFound(w)
			return
		}

		// Otherwise, there's some error with the database itself.
		logger.Error("error getdel-ing from the database: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token := c.Val()
	decrypted, err := Decrypt(token, key)
	if err != nil {
		logger.Error("error decrypting the secret", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := view.ShowPassword(w, decrypted); err != nil {
		logger.Error("error rendering ShowPassword view", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func splitToken(t string) (string, string, error) {
	spl := strings.Split(t, tokenSeparator)
	if len(spl) != 2 {
		return "", "", fmt.Errorf("unable to split token: '%s' using separator '%s'", t, tokenSeparator)
	}

	return spl[0], spl[1], nil
}
