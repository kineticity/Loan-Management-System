package modules

import (
	"loanApp/app"
	"loanApp/components/customer/controller"
	"loanApp/components/customer/service"
)

func registerCustomerRoutes(appObj *app.App) {
	customerService := service.NewCustomerService(appObj.DB, appObj.Repository, appObj.Log)
	customerController := controller.NewCustomerController(customerService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{customerController})
}
