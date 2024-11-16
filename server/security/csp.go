package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/logging"
)

type cspNonceKey struct{}

func CSPNonce(ctx context.Context) (string, error) {
	nonce, ok := ctx.Value(cspNonceKey{}).(string)
	if !ok {
		return "", fmt.Errorf("no nonce in context")
	}
	return nonce, nil
}

// isDevStaticFileRequest returns true if the request is for a static HTML file.
//
// We need to relax the CSP rules for these files because they can contain inline scripts.
//
// This can only really happen in development mode when serving uploaded HTML objects from the app.
// In prod, this isn't an issue because these files are not served by the app.
func isDevStaticFileRequest(conf *config.Config, r *http.Request) bool {
	return conf.DevMode && strings.HasPrefix(r.URL.Path, "/dev/storage/html/")
}

func NewCSPMiddleware(conf *config.Config, logger logging.Logger, next http.Handler) http.Handler {
	u := *conf.FileURLPattern
	u.Path = ""
	fileURLBase := u.String()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		csp := strings.Builder{}

		// default-src
		fmt.Fprintf(&csp, "default-src 'self' %s", fileURLBase)
		if isDevStaticFileRequest(conf, r) {
			// Needed for embedded images in rendered markdown dev pastes.
			fmt.Fprintf(&csp, " *")
		}

		// script-src
		fmt.Fprintf(&csp, "; script-src https://ajax.googleapis.com %s", fileURLBase)
		if isDevStaticFileRequest(conf, r) {
			fmt.Fprintf(&csp, " 'unsafe-inline'")
		} else {
			nonceBytes := make([]byte, 16)
			if _, err := rand.Read(nonceBytes); err != nil {
				logger.Error(ctx, "generating nonce", "error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			nonce := hex.EncodeToString(nonceBytes)
			ctx = context.WithValue(ctx, cspNonceKey{}, nonce)
			fmt.Fprintf(&csp, " 'nonce-%s'", nonce)
		}

		// style-src
		fmt.Fprintf(&csp, "; style-src 'self' https://fonts.googleapis.com %s", fileURLBase)

		// font-src
		fmt.Fprintf(&csp, "; font-src https://fonts.gstatic.com %s", fileURLBase)
		w.Header().Set("Content-Security-Policy", csp.String())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
