package web

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/unluckythoughts/go-microservice/v2/utils"
)

const (
	csrfTokenSessionKey = "csrf_token"
	// CSRFTokenHeader is the HTTP header name used to pass CSRF tokens in requests.
	CSRFTokenHeader = "X-CSRF-Token"
	// CSRFTokenFormField is the HTML form field name used to pass CSRF tokens.
	CSRFTokenFormField = "_csrf_token"
)

// GenerateCSRFToken creates a cryptographically random CSRF token, persists it in
// the session, and returns the token string.  Call this after a user authenticates
// and include the returned value in the login response so the client can supply it
// on subsequent state-changing requests.
func GenerateCSRFToken(ctx Context) (string, error) {
	token, err := utils.GenerateRandomString(64)
	if err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	if err := ctx.PutSessionValue(csrfTokenSessionKey, token); err != nil {
		return "", fmt.Errorf("failed to store CSRF token in session: %w", err)
	}

	return token, nil
}

// ValidateCSRFToken checks that the CSRF token in the request matches the one
// stored in the session.  Call this on state-changing requests when the caller
// has already established that the user is authenticated via a session cookie.
// The token is read from the X-CSRF-Token header; the _csrf_token form field is
// accepted as a fallback for HTML form submissions.
func ValidateCSRFToken(r Request) error {
	storedRaw, err := r.GetContext().GetSessionValue(csrfTokenSessionKey)
	if err != nil {
		return NewError(http.StatusForbidden, fmt.Errorf("CSRF token not found in session"))
	}

	stored, ok := storedRaw.(string)
	if !ok || stored == "" {
		return NewError(http.StatusForbidden, fmt.Errorf("CSRF token not found in session"))
	}

	provided := r.GetHeader(CSRFTokenHeader)
	if provided == "" {
		provided = r.GetHeaders().Get(CSRFTokenFormField)
	}

	if subtle.ConstantTimeCompare([]byte(stored), []byte(provided)) != 1 {
		return NewError(http.StatusForbidden, fmt.Errorf("CSRF token validation failed"))
	}

	return nil
}
