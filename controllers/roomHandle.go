package controllers

import (
	// "bilibili-recording-go/live"
	"bilibili-recording-go/tools"
	"encoding/json"
	"fmt"

	beego "github.com/beego/beego/v2/server/web"
	// "github.com/kataras/golog"
)

type RoomHandleController struct {
	beego.Controller
}

type receiveInfo struct {
	Handle	string		`json:"handle"`
	Data	roomInfo	`json:"data"`
}

type roomInfo struct {
	RoomID	string	`json:"RoomID"`
	RecordMode	bool	`json:"RecordMode"`
	StartTime	string	`json:"StartTime"`
	EndTime		string	`json:"EndTime"`
	AutoRecord	bool	`json:"AutoRecord"`
	AutoUpload	bool	`json:"AutoUpload"`
}

func (c *RoomHandleController) Post() {
	// golog.Info("Http Post at index.html")
	// c.Data["json"] = live.Lives
	// c.ServeJSON()
	fmt.Println(tools.BytesToStringFast(c.Ctx.Input.RequestBody))
	info := new(receiveInfo)
	json.Unmarshal(c.Ctx.Input.RequestBody, info)
	fmt.Println(info)
	//TODO: 房间的编辑删除增加逻辑
}