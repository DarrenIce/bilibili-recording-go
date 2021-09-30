package routers

import (
	"bilibili-recording-go/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
    beego.Router("/", &controllers.MainController{})
	beego.Router("/live-info", &controllers.LiveController{})
	beego.Router("/base-info", &controllers.BaseController{})
	beego.Router("/room-handle", &controllers.RoomHandleController{})
}
