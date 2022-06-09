package controllers

import (
	"fmt"
	"time"

	"bilibili-recording-go/live"

	"github.com/gin-gonic/gin"
	"github.com/kataras/golog"
)


type receiveInfo2 struct {
	RoomID string   `json:"roomID"`
}


func ProcessDecode(c *gin.Context) {
	info := receiveInfo2{}
	if c.ShouldBind(&info) == nil {
		fmt.Println(info)
		if room, ok := live.Lives[info.RoomID]; ok {
			if room.State != 5 {
				golog.Debug(fmt.Sprintf("%s[RoomID: %s] 开始转码", room.Uname, info.RoomID))
				room.DecodeStartTime = time.Now().Unix()
				room.Decode()
				golog.Debug(fmt.Sprintf("%s[RoomID: %s] 结束转码", room.Uname, info.RoomID))
				room.DecodeEndTime = time.Now().Unix()
				c.JSON(200, &struct {
					Msg bool `json:"msg"`
				}{
					true,
				})
			} else {
				c.JSON(200, &struct {
					Msg bool `json:"msg"`
				}{
					false,
				})
			}
		} else {
			c.JSON(200, &struct {
				Msg bool `json:"msg"`
			}{
				false,
			})
		}
	}
}