package controller

import (
	"encoding/json"
	"net/http"
	"time"

	loginservice "loanApp/components/login/service" 
	"loanApp/components/middleware"                 
	"loanApp/components/user/service"              
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

	user, err := lc.UserService.AuthenticateUser(credentials.Email, credentials.Password)
	if err != nil {
		web.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	expirationTime := time.Now().Add(1 * time.Hour) // Token expires in 1 hour
	claims := middleware.NewClaims(user.ID, user.Email, user.Role, expirationTime)

	token, err := claims.Signing()
	if err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create token")
		return
	}

	info := &logininfo.LoginInfo{UserID: user.ID, LoginTime: time.Now()}
	lc.LoginService.CreateLoginInfo(user, info)

	web.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token":     token,
		"expiresAt": claims.ExpiresAt,
		"role":      claims.Role,
		"user_id":   claims.UserID,
	})
}
