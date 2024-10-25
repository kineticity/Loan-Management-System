package controller

import (
	"fmt"
	"net/http"
	"strings"

	logoutservice "loanApp/components/logout/service" // Adjust based on your project structure
	"loanApp/components/middleware"                   // Import your middleware package
	"loanApp/components/user/service"                 // Adjust based on your project structure

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

type LogoutController struct {
	UserService   *service.UserService
	LogoutService *logoutservice.LogoutService
}

func NewLogoutController(userService *service.UserService, logoutService *logoutservice.LogoutService) *LogoutController {
	return &LogoutController{
		UserService:   userService,
		LogoutService: logoutService,
	}
}

// RegisterRoutes registers the login route
func (lc *LogoutController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/logout", lc.LogoutHandler).Methods(http.MethodPost)
}
func (lc *LogoutController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Extract token from the Authorization header
	tok := r.Header.Get("Authorization")
	if tok == "" {
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		return
	}

	// Remove the "Bearer " prefix
	tokenStr := strings.TrimPrefix(tok, "Bearer ")

	// Check if the token is already blacklisted
	if middleware.IsTokenBlacklisted(tokenStr) {
		http.Error(w, "Token has already been invalidated", http.StatusUnauthorized)
		return
	}

	// Parse the token
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the token's signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("it'sDevthedev"), nil // Replace with your actual secret key
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Retrieve the user by email from the claims
	email := claims.Email
	user, err := lc.UserService.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusUnauthorized)
		return
	}

	// Blacklist the token to prevent further use
	middleware.BlacklistToken(tokenStr)

	// Update the logout information for the user
	if err := lc.LogoutService.UpdateLoginInfo(user); err != nil {
		http.Error(w, "Failed to update logout info", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}
