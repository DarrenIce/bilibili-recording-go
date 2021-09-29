package controllers

import (
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