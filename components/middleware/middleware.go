package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"loanApp/models/user"
	"loanApp/models/userclaims"
	"loanApp/utils/web"

	"github.com/dgrijalva/jwt-go"
)

var SecretKey = []byte("it'sDevthedev")

// SigningToken signs the JWT token with the user information
func SigningToken(u *user.User, expirationTime time.Time) (string, error) {
	claims := &userclaims.UserClaims{
		User: *u, // Embed the user struct
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}

// TokenAuthMiddleware checks the JWT token in the Authorization header
func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if IsTokenBlacklisted(tokenStr) {
			http.Error(w, "Token has been invalidated, please log in again", http.StatusUnauthorized)
			return
		}

		claims, err := VerifyJWT(tokenStr)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// VerifyJWT verifies the token and extracts the claims
func VerifyJWT(tokenStr string) (*userclaims.UserClaims, error) {
	claims := &userclaims.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// AdminOnly middleware restricts access to admins only
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value("claims").(*userclaims.UserClaims)
		if !ok || claims == nil {
			web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: claims not found")
			return
		}

		if claims.User.Role != "Admin" {
			web.RespondWithError(w, http.StatusForbidden, "Unauthorized access")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CustomerOnly middleware restricts access to customers only
func CustomerOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value("claims").(*userclaims.UserClaims)
		if !ok || claims == nil {
			web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: claims not found")
			return
		}

		if claims.User.Role != "Customer" {
			web.RespondWithError(w, http.StatusForbidden, "Unauthorized access")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CustomerOnly middleware restricts access to customers only
func LoanOfficerOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value("claims").(*userclaims.UserClaims)
		if !ok || claims == nil {
			web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: claims not found")
			return
		}

		if claims.User.Role != "Loan Officer" {
			web.RespondWithError(w, http.StatusForbidden, "Unauthorized access")
			return
		}

		next.ServeHTTP(w, r)
	})
}
