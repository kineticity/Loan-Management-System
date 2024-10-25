package modules

import (
	"loanApp/app"
	"loanApp/components/logout/controller"
	"loanApp/components/user/service"
	logoutservice "loanApp/components/logout/service"

)

func registerLogoutRoutes(appObj *app.App) {

	userService := service.NewUserService(appObj.DB, appObj.Repository, appObj.Log)
	logoutService := logoutservice.NewLogoutService(appObj.DB, appObj.Repository, appObj.Log)

	logoutController:=controller.NewLogoutController(userService,logoutService)
	// adminController := controller.NewAdminController(adminService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{logoutController})
}
