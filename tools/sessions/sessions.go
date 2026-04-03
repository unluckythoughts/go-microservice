package sessions

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/unluckythoughts/go-microservice/v2/tools/web"
	"github.com/unluckythoughts/go-microservice/v2/utils"
	"go.uber.org/zap"
)

var (
	// Default session cookie name
	sessionName = "session"
)

type (
	// Store interface for session management
	// TODO: add redis store to save sessions
	Store interface {
		// Get retrieves a session for the given request
		Get(req *http.Request, key string) (*sessions.Session, error)
		// Save saves the session for the given request and response writer
		Save(req *http.Request, resp http.ResponseWriter, value *sessions.Session) error
		// New creates a new session for the given request
		New(req *http.Request, key string) (*sessions.Session, error)
	}

	// Options contains configuration for session management
	Options struct {
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
		HttpOnly bool `env:"SESSION_HTTP_ONLY" envDefault:"false"`
		// SameSite determines the SameSite cookie attribute
		SameSite http.SameSite `env:"SESSION_SAME_SITE" envDefault:"1"`
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

// NewStore creates a new session store with the given options
func NewStore(opts Options) Store {
	var secretKey []byte

	if opts.SecretKey != "" {
		secretKey = []byte(opts.SecretKey)
	} else {
		strKey, err := utils.GenerateRandomString(32)
		if err != nil {
			panic(err)
		}
		secretKey = []byte(strKey)
		opts.Logger.Info("Generated random session secret key")
		opts.Logger.Debug("Session secret initialized", zap.Int("length", len(secretKey)))
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

// GetMiddleware creates a middleware that injects session management into the request context
func GetMiddleware(store Store) web.Middleware {
	return func(req web.MiddlewareRequest) error {
		session, err := store.Get(req.GetInternalRequest(), sessionName)
		if err != nil {
			// Gorilla sessions always returns a valid session object alongside a decode
			// error (e.g. "securecookie: the value is not valid").  Calling store.New()
			// as a fallback hits the same invalid cookie and produces the same error.
			// Instead, reset the returned session to a clean state and carry on so the
			// bad cookie is silently replaced at the end of the request.
			if session == nil {
				return err
			}
			session.Values = make(map[any]any)
			session.IsNew = true
		}

		// Set the session in the request context
		req.GetContext().SetSession(session)

		return nil
	}
}
