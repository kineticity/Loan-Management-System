package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"loanApp/models/user" // Ensure to import your user model package
	"loanApp/utils/web"

	"github.com/golang-jwt/jwt"
)

var SecretKey = []byte("it'sDevthedev") // Make sure to use a secure key in production

// Claims struct to hold the user claims
type Claims struct {
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Role     user.Role `json:"role"`
	jwt.StandardClaims
}

// NewClaims creates a new Claims object
func NewClaims(email string, password string, role user.Role, expirationDate time.Time) *Claims {
	return &Claims{
		Email:    email,
		Password: password,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationDate.Unix(),
		},
	}
}

// Valid checks if the token has expired
func (c *Claims) Valid() error {
	if time.Now().Unix() > c.ExpiresAt {
		return errors.New("token has expired")
	}
	return nil
}

// Signing creates a signed token string from the claims
func (c *Claims) Signing() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(SecretKey)
}

// TokenAuthMiddleware checks for a valid JWT token
func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		claims, err := VerifyJWT(authHeader)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// VerifyJWT verifies the JWT token and returns the claims
func VerifyJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
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

// VerifyRoleMiddleware checks if the user has one of the allowed roles
func VerifyRoleMiddleware(allowedRoles ...user.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("claims").(*Claims)
			if !ok || claims == nil {
				http.Error(w, "Unauthorized: claims not found", http.StatusUnauthorized)
				return
			}

			isAllowed := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				http.Error(w, "Unauthorized: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminOnly is middleware to allow only admins to access certain routes
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract claims from context
		claims, ok := r.Context().Value("claims").(*Claims)
		if !ok || claims == nil {
			web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: claims not found")
			return
		}

		// Check if the role is "Admin"
		if claims.Role != "Admin" {
			web.RespondWithError(w, http.StatusForbidden, "Unauthorized access")
			return
		}

		// If the user is an admin, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}