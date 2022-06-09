package routers

import (
	"bilibili-recording-go/controllers"

	"github.com/gin-gonic/gin"
)

var GIN *gin.Engine

func init() {
    GIN = gin.Default()
	GIN.GET("/base", controllers.GetBaseStatus)
}
