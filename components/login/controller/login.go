package controller

import (
	"encoding/json"
	"net/http"
	"time"

	loginservice "loanApp/components/login/service" // Adjust based on your project structure
	"loanApp/components/middleware"                 // Import your middleware package
	"loanApp/components/user/service"               // Adjust based on your project structure
	"loanApp/models/logininfo"

	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type LoginController struct {
	UserService  *service.UserService
	LoginService *loginservice.LoginService
}

func NewLoginController(userService *service.UserService, loginService *loginservice.LoginService) *LoginController {
	return &LoginController{
		UserService:  userService,
		LoginService: loginService,
	}
}

// RegisterRoutes registers the login route
func (lc *LoginController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", lc.Login).Methods(http.MethodPost)
}

func (lc *LoginController) Login(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		web.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Authenticate the user
	user, err := lc.UserService.AuthenticateUser(credentials.Email, credentials.Password)
	if err != nil {
		web.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Create JWT claims with user ID and role
	expirationTime := time.Now().Add(1 * time.Hour) // Token expires in 1 hour
	claims := middleware.NewClaims(user.ID, user.Email, user.Role, expirationTime)

	// Sign the token
	token, err := claims.Signing()
	if err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create token")
		return
	}

	// Store login information
	info := &logininfo.LoginInfo{UserID: int(user.ID), LoginTime: time.Now()}
	lc.LoginService.CreateLoginInfo(user, info)

	// Respond with the token and other information
	web.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token":     token,
		"expiresAt": claims.ExpiresAt,
		"role":      claims.Role,
		"user_id":   claims.UserID,
	})
}
