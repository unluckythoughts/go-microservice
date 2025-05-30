package web

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

var (
	// Default session cookie name
	sessionName = "session"
)

type (
	// SessionStore interface for session management
	// TODO: add redis store to save sessions
	SessionStore interface {
		// Get retrieves a session for the given request
		Get(req *http.Request, key string) (*sessions.Session, error)
		// Save saves the session for the given request and response writer
		Save(req *http.Request, resp http.ResponseWriter, value *sessions.Session) error
		// New creates a new session for the given request
		New(req *http.Request, key string) (*sessions.Session, error)
	}

	// SessionOptions contains configuration for session management
	SessionOptions struct {
		// Name is the name of the session cookie
		Name string `env:"SESSION_NAME" envDefault:"session"`
		// SecretKey is the key used to sign the session cookie
		// If not provided, a random key will be generated
		SecretKey string `env:"SESSION_SECRET_KEY"`
		// MaxAge is the maximum age of the session cookie in seconds
		MaxAge int `env:"SESSION_MAX_AGE" envDefault:"86400"`
		// Secure determines if the cookie should only be sent over HTTPS
		Secure bool `env:"SESSION_SECURE" envDefault:"false"`
		// HttpOnly determines if the cookie should only be accessible via HTTP(S)
		HttpOnly bool `env:"SESSION_HTTP_ONLY" envDefault:"true"`
		// SameSite determines the SameSite cookie attribute
		SameSite http.SameSite `env:"SESSION_SAME_SITE" envDefault:"Lax"`
		// Logger is the logger for the session store
		Logger *zap.Logger
	}

	sessionStore struct {
		store *sessions.CookieStore
		l     *zap.Logger
	}
)

func (s *sessionStore) New(req *http.Request, key string) (*sessions.Session, error) {
	return s.store.New(req, key)
}

// Get retrieves a session for the given request
func (s *sessionStore) Get(req *http.Request, key string) (*sessions.Session, error) {
	return s.store.Get(req, key)
}

// Save saves the session for the given request and response writer
func (s *sessionStore) Save(req *http.Request, resp http.ResponseWriter, value *sessions.Session) error {
	return s.store.Save(req, resp, value)
}

// generateRandomKey generates a random 32-byte key for session encryption
func generateRandomKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// NewSessionStore creates a new session store with the given options
func NewSessionStore(opts SessionOptions) SessionStore {
	var secretKey []byte
	var err error

	if opts.SecretKey != "" {
		secretKey = []byte(opts.SecretKey)
	} else {
		secretKey, err = generateRandomKey()
		if err != nil {
			panic(err)
		}
		opts.Logger.Info("Generated random session secret key", zap.String("key", base64.StdEncoding.EncodeToString(secretKey)))
	}

	if opts.Name != "" {
		sessionName = opts.Name
	}

	// Default session cookie max age (24 hours)
	sessionMaxAge := 24 * 60 * 60
	if opts.MaxAge != 0 {
		sessionMaxAge = opts.MaxAge
	}

	store := sessions.NewCookieStore(secretKey)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionMaxAge,
		Secure:   opts.Secure,
		HttpOnly: opts.HttpOnly,
		SameSite: opts.SameSite,
	}

	return &sessionStore{
		store: store,
		l:     opts.Logger,
	}
}

// SessionMiddleware creates a middleware that injects session management into the request context
func SessionMiddleware(store SessionStore) Middleware {
	return func(req MiddlewareRequest) error {
		session, err := store.Get(req.(*request)._int, sessionName)
		if err != nil {
			return err
		}

		// Set the session in the request context
		req.GetContext().SetSession(session)

		return nil
	}
}
