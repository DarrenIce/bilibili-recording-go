package controllers

import (
	"bilibili-recording-go/monitor"

	beego "github.com/beego/beego/v2/server/web"
)

type MonitorController struct {
	beego.Controller
}

func (c *MonitorController) Get() {
	c.Data["json"] = monitor.AreaMonitorMap
	c.ServeJSON()
}
