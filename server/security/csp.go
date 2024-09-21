package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/chriskuehl/fluffy/server/config"
)

type cspNonceKey struct{}

func CSPNonce(ctx context.Context) (string, error) {
	nonce, ok := ctx.Value(cspNonceKey{}).(string)
	if !ok {
		return "", fmt.Errorf("no nonce in context")
	}
	return nonce, nil
}

func NewCSPMiddleware(conf *config.Config, next http.Handler) http.Handler {
	fileURLBase := *conf.FileURLPattern
	fileURLBase.Path = ""
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		nonceBytes := make([]byte, 16)
		if _, err := rand.Read(nonceBytes); err != nil {
			panic("failed to generate nonce: " + err.Error())
		}
		nonce := hex.EncodeToString(nonceBytes)
		ctx = context.WithValue(ctx, cspNonceKey{}, nonce)
		csp := fmt.Sprintf(
			"default-src 'self' %s; script-src https://ajax.googleapis.com 'nonce-%s' %[1]s; style-src 'self' https://fonts.googleapis.com %[1]s; font-src https://fonts.gstatic.com %[1]s",
			fileURLBase.String(),
			nonce,
		)
		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
