package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"bilibili-recording-go/live"
	"bilibili-recording-go/tools"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/kataras/golog"
)

type DecodeController struct {
	beego.Controller
}

type receiveInfo2 struct {
	RoomID string   `json:"RoomID"`
}


func (c *DecodeController) Post() {
	fmt.Println(tools.BytesToStringFast(c.Ctx.Input.RequestBody))
	info := new(receiveInfo2)
	json.Unmarshal(c.Ctx.Input.RequestBody, info)
	fmt.Println(info)
	if room, ok := live.Lives[info.RoomID]; ok {
		if room.State != 5 {
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 开始转码", room.Uname, info.RoomID))
			room.DecodeStartTime = time.Now().Unix()
			room.Decode()
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 结束转码", room.Uname, info.RoomID))
			room.DecodeEndTime = time.Now().Unix()
			c.Data["json"] = &struct {
				Msg bool `json:"msg"`
			}{
				true,
			}
		} else {
			c.Data["json"] = &struct {
				Msg bool `json:"msg"`
			}{
				false,
			}
		}
	} else {
		c.Data["json"] = &struct {
			Msg bool `json:"msg"`
		}{
			false,
		}
	}
	c.ServeJSON()
}