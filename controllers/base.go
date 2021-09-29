package controllers

import (
	"bilibili-recording-go/tools"

	beego "github.com/beego/beego/v2/server/web"
	// "github.com/kataras/golog"
)

type BaseController struct {
	beego.Controller
}


func (c *BaseController) Post() {
	// golog.Info("Http Post at index.html")
	c.Data["json"] = struct {
		TotalDownload	int64
		FileNum			int64

	} {
		tools.DirSize("./recording", 0),
		tools.CacRecordingFileNum(),
	}
	c.ServeJSON()
}