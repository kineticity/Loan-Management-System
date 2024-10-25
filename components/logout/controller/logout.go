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

	// Parse the token
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure to validate the token's signing method here
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key (replace with your actual secret)
		return []byte("it'sDevthedev"), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Now you can get the user ID from the claims
	email := claims.Email
	user, err := lc.UserService.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusUnauthorized)
		return
	}

	// Blacklist the token
	middleware.BlacklistToken(tok)
	lc.LogoutService.UpdateLoginInfo(user)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}
