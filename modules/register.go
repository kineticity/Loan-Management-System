package modules

import (
	"loanApp/app"
	"loanApp/components/register/controller"
	"loanApp/components/customer/service"

)

func registerRegisterRoutes(appObj *app.App) {

	customerService := service.NewCustomerService(appObj.DB, appObj.Repository, appObj.Log)
	// loginService := loginservice.NewLoginService(appObj.DB, appObj.Repository, appObj.Log)

	loginController:=controller.NewRegisterController(customerService,appObj.Log)
	// adminController := controller.NewAdminController(adminService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{loginController})
}
