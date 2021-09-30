package controllers

import (
	"bilibili-recording-go/live"

	beego "github.com/beego/beego/v2/server/web"
	// "github.com/kataras/golog"
)

type RoomHandleController struct {
	beego.Controller
}

func (c *RoomHandleController) Post() {
	// golog.Info("Http Post at index.html")
	c.Data["json"] = live.Lives
	c.ServeJSON()
}