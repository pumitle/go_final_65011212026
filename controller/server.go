package controller

import "github.com/gin-gonic/gin"

func StartServer() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	CustomerController(router)
	ProductController(router)
	router.Run()
}
