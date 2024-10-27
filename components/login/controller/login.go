package controller

import (
	"encoding/json"
	"loanApp/models/user"
	"net/http"
	"time"

	loginservice "loanApp/components/login/service"
	"loanApp/components/middleware"
	"loanApp/components/user/service"
	"loanApp/models/logininfo"

	"loanApp/utils/web"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
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

	// Check if the user is already logged in
	activeLogin, err := lc.LoginService.GetActiveLoginInfo(int(user.ID))
	if err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, "Error checking active login")
		return
	}

	// Token expiration time
	expirationTime := time.Now().Add(1 * time.Hour)

	// If the user is already logged in, respond with the existing token
	if activeLogin != nil {
		token, err := lc.createToken(user, expirationTime)
		if err != nil {
			web.RespondWithError(w, http.StatusInternalServerError, "Error generating token")
			return
		}

		// Respond with existing login token
		web.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"message":   "User already logged in",
			"token":     token,
			"expiresAt": expirationTime,
			"role":      user.Role,
			"user_id":   user.ID,
		})
		return
	}

	// No active session found, proceed with creating new login session
	token, err := lc.createToken(user, expirationTime)
	if err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create token")
		return
	}

	// Store login information
	info := &logininfo.LoginInfo{UserID: user.ID, LoginTime: time.Now()}
	err = lc.LoginService.CreateLoginInfo(user, info)
	if err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, "Error creating login info")
		return
	}

	// Respond with the token and other information
	web.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token":     token,
		"expiresAt": expirationTime,
		"role":      user.Role,
		"user_id":   user.ID,
	})
}

// createToken generates a token for the authenticated user
func (lc *LoginController) createToken(u *user.User, expirationTime time.Time) (string, error) {
	// Create the token claims using the user struct directly
	claims := &user.User{
		Model:    gorm.Model{ID: u.ID},
		Email:    u.Email,
		Role:     u.Role,
		IsActive: u.IsActive,
	}

	// Generate the token using the claims
	token, err := middleware.SigningToken(claims, expirationTime) // Adjust this function to handle user.User
	if err != nil {
		return "", err
	}

	return token, nil
}
