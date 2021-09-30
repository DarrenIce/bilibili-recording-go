package controllers

import (
	// "bilibili-recording-go/live"
	// "bilibili-recording-go/config"
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
	Handle string   `json:"handle"`
	Data   roomInfo `json:"data"`
}

type roomInfo struct {
	RoomID         string `json:"RoomID"`
	RecordMode     bool   `json:"RecordMode"`
	StartTime      string `json:"StartTime"`
	EndTime        string `json:"EndTime"`
	AutoRecord     bool   `json:"AutoRecord"`
	AutoUpload     bool   `json:"AutoUpload"`
	NeedM4a        bool   `json:"NeedM4a"`
	Mp4Compress    bool   `json:"Mp4Compress"`
	DivideByTitle  bool   `json:"DivideByTitle"`
	CleanUpRegular bool   `json:"CleanUpRegular"`
	SaveDuration   string `json:"SaveDuration"`
	AreaLock       bool   `json:"AreaLock"`
	AreaLimit      string `json:"AreaLimit"`
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
	if info.Handle == "add" {
		c.Data["json"] = &struct {
			Msg bool `json:"msg"`
		}{
			addRoom(info.Data),
		}
	} else if info.Handle == "edit" {
		c.Data["json"] = &struct {
			Msg bool `json:"msg"`
		}{
			editRoom(info.Data),
		}
	} else if info.Handle == "delete" {
		c.Data["json"] = &struct {
			Msg bool `json:"msg"`
		}{
			deleteRoom(info.Data),
		}
	}
	c.ServeJSON()
}

func addRoom(info roomInfo) bool {
	// c := config.New()
	// c.AddRoom(config.RoomConfigInfo{
	// 	RoomID:         info.RoomID,
	// 	StartTime:      info.StartTime,
	// 	EndTime:        info.EndTime,
	// 	AutoRecord:     info.AutoRecord,
	// 	AutoUpload:     info.AutoUpload,
	// 	RecordMode:     info.RecordMode,
	// 	NeedM4a:        info.NeedM4a,
	// 	Mp4Compress:    info.Mp4Compress,
	// 	DivideByTitle:  info.DivideByTitle,
	// 	CleanUpRegular: info.CleanUpRegular,
	// 	SaveDuration:   info.SaveDuration,
	// 	AreaLock:       info.AreaLock,
	// 	AreaLimit:      info.AreaLimit,
	// })
	return true
}

func editRoom(info roomInfo) bool {
	return true
}

func deleteRoom(info roomInfo) bool {
	return true
}
