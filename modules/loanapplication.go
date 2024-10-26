package modules

import (
	"loanApp/app"
	"loanApp/components/loanapplication/controller"
	"loanApp/components/loanapplication/service"
)

func registerLoanApplicationRoutes(appObj *app.App) {
	loanofficerService := service.NewLoanApplicationService(appObj.DB, appObj.Repository, appObj.Log)
	loanofficerController := controller.NewLoanApplicationController(loanofficerService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{loanofficerController})
}
