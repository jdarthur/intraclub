package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/controllers"
	"intraclub/middleware"
	"intraclub/model"
)

func main() {
	fmt.Println("Intraclub")

	common.GlobalDbProvider = model.NewMongoDbProvider("", "", "")
	router := gin.Default()

	apiAuthAndAccess := router.Group("/api", middleware.AuthCheck())

	userCtl := common.CrudController{Controller: controllers.UserController{}, Database: common.GlobalDbProvider}
	apiAuthAndAccess.Handle("GET", "/api/users", userCtl.GetAll)
	apiAuthAndAccess.Handle("GET", "/api/users/:id", userCtl.GetOne)
	apiAuthAndAccess.Handle("POST", "/api/users", userCtl.Create)
	apiAuthAndAccess.Handle("PUT", "/api/users/:id", userCtl.Update)
	apiAuthAndAccess.Handle("DELETE", "/api/users/:id", userCtl.Delete)
}
