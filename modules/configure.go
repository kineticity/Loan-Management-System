package modules

import (
	"loanApp/app"

	"loanApp/models/logininfo"
	"loanApp/models/user"
)

func ConfigureAppTables(appObj *app.App) {
	userModuleConfig := user.NewUserModuleConfig(appObj.DB, appObj.Log)
	loginModuleConfig := logininfo.NewLoginInfoModuleConfig(appObj.DB, appObj.Log)

	appObj.TableMigration([]app.ModuleConfig{userModuleConfig, loginModuleConfig})
}
