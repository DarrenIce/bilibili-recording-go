package controllers

import (
	"bilibili-recording-go/live"

	beego "github.com/beego/beego/v2/server/web"
	// "github.com/kataras/golog"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	// golog.Info("Http Get at index.html")
	c.TplName = "index.html"
}

func (c *MainController) Post() {
	// golog.Info("Http Post at index.html")
	c.Data["json"] = live.Lives
	c.ServeJSON()
}