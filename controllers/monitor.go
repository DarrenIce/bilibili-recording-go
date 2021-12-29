package controllers

import (
	"bilibili-recording-go/monitor"

	beego "github.com/beego/beego/v2/server/web"
)

type MonitorController struct {
	beego.Controller
}

func (c *MonitorController) Post() {
	c.Data["json"] = monitor.MonitorMap
	c.ServeJSON()
}