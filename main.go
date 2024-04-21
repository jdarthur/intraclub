package main

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/controllers"
	"intraclub/middleware"
	"intraclub/model"
	"sync"
)

func main() {
	common.GlobalDbProvider = model.NewMongoDbProvider("mongodb://localhost:27018", "", "")
	err := common.GlobalDbProvider.Connect()
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	noAuth := router.Group("/api")

	otpManager := controllers.OneTimePasswordManager{Map: &sync.Map{}}
	noAuth.POST("/token", otpManager.GetToken)
	noAuth.POST("/one_time_password", otpManager.Create)

	registerManager := controllers.NewUserController{}
	noAuth.POST("/register", registerManager.Register)

	apiAuthAndAccess := router.Group("/api", middleware.WithToken)

	userCtl := common.CrudController{Controller: controllers.UserController{}, Database: common.GlobalDbProvider}
	apiAuthAndAccess.Handle("GET", "/users", userCtl.GetAll)
	apiAuthAndAccess.Handle("GET", "/users/:id", userCtl.GetOne)

	// you can only delete your own user
	ownedByUser := middleware.OwnedByUserWrapper{Record: &model.User{}}
	apiAuthAndAccess.Handle("DELETE", "/users/:id", ownedByUser.OwnedByUser, userCtl.Delete)

	leagueCtl := common.CrudController{Controller: controllers.LeagueController{}, Database: common.GlobalDbProvider}
	apiAuthAndAccess.Handle("GET", "/users", userCtl.GetAll)
	apiAuthAndAccess.Handle("GET", "/users/:id", userCtl.GetOne)

	err = router.Run(":8080")
	if err != nil {
		panic(err)
	}

}
