package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dockyguy56/ecomm/internal/token"
)

type authKey struct{}

func GetAuthMiddlewareFnuc(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read auth header
			// verify the token
			claims, err := verifyClaimsfromAuthHeader(r, tokenMaker)
			if err != nil {
				http.Error(w, fmt.Sprint("error verifying token: %v", err), http.StatusUnauthorized)
				return
			}
			// pass the payload/claims down the context
			ctx := context.WithValue(r.Context(), authKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAdminMiddlewareFnuc(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read auth header
			// verify the token
			claims, err := verifyClaimsfromAuthHeader(r, tokenMaker)
			if err != nil {
				http.Error(w, fmt.Sprint("error verifying token: %v", err), http.StatusUnauthorized)
				return
			}

			if !claims.IsAdmin {
				http.Error(w, "user is not admin", http.StatusForbidden)
				return
			}
			// pass the payload/claims down the context
			ctx := context.WithValue(r.Context(), authKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func verifyClaimsfromAuthHeader(r *http.Request, tokenMaker *token.JWTMaker) (*token.UserClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("Authorization heaser is missing")
	}

	fields := strings.Fields(authHeader) // Bearer <token>
	if len(fields) != 2 || fields[0] != "Bearer" {
		return nil, fmt.Errorf("Invalid authorization header")
	}

	token := fields[1]
	claims, err := tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, fmt.Errorf("Invalid token: %w", err)
	}

	return claims, nil
}
