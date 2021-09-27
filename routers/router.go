package routers

import (
	"bilibili-recording-go/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}
