package modules

import (
	"loanApp/app"
	"loanApp/components/login/controller"
	"loanApp/components/user/service"
	loginservice "loanApp/components/login/service"

)

func registerLoginRoutes(appObj *app.App) {

	userService := service.NewUserService(appObj.DB, appObj.Repository, appObj.Log)
	loginService := loginservice.NewLoginService(appObj.DB, appObj.Repository, appObj.Log)

	loginController:=controller.NewLoginController(userService,loginService)
	// adminController := controller.NewAdminController(adminService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{loginController})
}
