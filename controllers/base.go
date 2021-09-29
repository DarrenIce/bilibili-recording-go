package controllers

import (
	"bilibili-recording-go/live"

	beego "github.com/beego/beego/v2/server/web"
	// "github.com/kataras/golog"
)

type BaseController struct {
	beego.Controller
}


func (c *BaseController) Post() {
	// golog.Info("Http Post at index.html")
	c.Data["json"] = live.Lives
	c.ServeJSON()
}