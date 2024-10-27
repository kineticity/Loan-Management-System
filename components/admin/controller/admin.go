package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"loanApp/components/admin/service"
	"loanApp/components/middleware"
	"loanApp/models/user"
	"loanApp/utils/log"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type AdminController struct {
	AdminService *service.AdminService
	log          log.Logger
}

func NewAdminController(AdminService *service.AdminService, log log.Logger) *AdminController {
	return &AdminController{
		AdminService: AdminService,
		log:          log,
	}
}

func (a *AdminController) RegisterRoutes(router *mux.Router) {
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.TokenAuthMiddleware)
	adminRouter.Use(middleware.AdminOnly)
	adminRouter.HandleFunc("/", a.CreateAdmin).Methods(http.MethodPost)
	adminRouter.HandleFunc("/", a.GetAllAdmins).Methods(http.MethodGet)
}

func (a *AdminController) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	a.log.Info("CreateAdmin called")

	var newAdmin user.Admin
	if err := json.NewDecoder(r.Body).Decode(&newAdmin); err != nil {
		a.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if err := validateAdmin(newAdmin); err != nil {
		a.log.Error("Validation error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	newAdmin.Role = "Admin" //<----------------------------------------------------------------------added

	if err := a.AdminService.CreateAdmin(&newAdmin); err != nil {
		a.log.Error("Error creating admin: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create admin")
		return
	}

	web.RespondWithJSON(w, http.StatusCreated, newAdmin)
}

func (a *AdminController) GetAllAdmins(w http.ResponseWriter, r *http.Request) {
	a.log.Info("GetAllAdmins called")

	parser := web.NewParser(r)
	allAdmins := []*user.Admin{}
	var totalCount int

	if err := a.AdminService.GetAllAdmins(&allAdmins, &totalCount, *parser); err != nil {
		a.log.Error("Error fetching admins: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch admins")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, allAdmins)
}

func validateAdmin(admin user.Admin) error {
	if admin.Name == "" {
		return errors.New("name cannot be empty")
	}
	if admin.Email == "" {
		return errors.New("email cannot be empty")
	}
	if admin.Password == "" {
		return errors.New("password cannot be empty")
	}

	return nil
}
