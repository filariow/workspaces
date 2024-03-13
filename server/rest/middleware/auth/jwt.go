package auth

import (
	"context"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"

	rcontext "github.com/konflux-workspaces/workspaces/server/core/context"
)

var _ http.Handler = &JwtBearerMiddleware{}

type JwtBearerMiddlewareBuilder struct {
	key []byte
}

func (b *JwtBearerMiddlewareBuilder) WithNext(handler http.Handler) *JwtBearerMiddleware {
	return NewJwtBearerMiddleware(b.key, handler)
}

type JwtBearerMiddleware struct {
	key  []byte
	next http.Handler
}

func NewJwtBearerMiddleware(validationKey []byte, next http.Handler) *JwtBearerMiddleware {
	return &JwtBearerMiddleware{key: validationKey, next: next}
}

func (p *JwtBearerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// extract bearer token from request
	t, ok := lookupBearerToken(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// parse token verifying signing key
	tkn, err := parseVerifiedToken(t, p.key)
	if err != nil {
		if _, err := w.Write([]byte(err.Error())); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// inject claims into request context
	nr, err := injectClaimsInRequestContext(r, tkn)
	if err != nil {
		if _, err := w.Write([]byte(err.Error())); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// serve next handler
	p.next.ServeHTTP(w, nr)
}

func injectClaimsInRequestContext(r *http.Request, tkn *jwt.Token) (*http.Request, error) {
	u, err := tkn.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, rcontext.UserKey, u)
	return r.WithContext(ctx), nil
}

func parseVerifiedToken(t string, signingKey []byte) (*jwt.Token, error) {
	jp := jwt.NewParser()
	sf := func(_ *jwt.Token) (interface{}, error) {
		return signingKey, nil
	}

	return jp.Parse(t, sf)
}

func lookupBearerToken(r *http.Request) (string, bool) {
	a := r.Header.Get("Authorization")
	t := strings.TrimPrefix(a, "Bearer ")
	return t, t != ""
}
