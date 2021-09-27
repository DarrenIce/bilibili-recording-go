package controllers

import (
	"fmt"

	"bilibili-recording-go/live"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/kataras/golog"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	c.Data["Lives"] = live.Lives
	c.TplName = "index.tpl"
	golog.Info("Http Get Lives")
}
