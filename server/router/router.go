package router

import (
	"context"
	"errors"
	"net/http"

	"github.com/offen/offen/server/persistence"
)

type router struct {
	db persistence.Database
}

type contextKey int

const (
	cookieKey                   = "user"
	contextKeyCookie contextKey = iota
)

func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/exchange":
		switch r.Method {
		case http.MethodGet:
			rt.getPublicKey(w, r)
		case http.MethodPost:
			rt.postUserSecret(w, r)
		default:
			respondWithError(w, errors.New("Method not allowed"), http.StatusMethodNotAllowed)
		}
	case "/events":
		c, err := r.Cookie(cookieKey)
		if err != nil {
			respondWithError(w, err, http.StatusBadRequest)
			return
		}
		if c.Value == "" {
			respondWithError(w, errors.New("received blank user identifier"), http.StatusBadRequest)
			return
		}

		r = r.WithContext(
			context.WithValue(r.Context(), contextKeyCookie, c.Value),
		)

		switch r.Method {
		case http.MethodGet:
			rt.getEvents(w, r)
		case http.MethodPost:
			rt.postEvents(w, r)
		default:
			respondWithError(w, errors.New("Method not allowed"), http.StatusMethodNotAllowed)
		}
	case "/status":
		rt.status(w, r)
	default:
		respondWithError(w, errors.New("Not found"), http.StatusNotFound)
	}
}

// New creates a new application router that reads and writes data
// to the given database implementation. In the context of the application
// this expects to be the only top level router in charge of handling all
// incoming HTTP requests.
func New(db persistence.Database) http.Handler {
	router := &router{db}
	withContentType := contentTypeMiddleware(router)
	withCors := corsMiddleware(withContentType)
	return withCors
}