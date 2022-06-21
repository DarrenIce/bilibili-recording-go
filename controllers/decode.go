package controllers

import (
	"fmt"

	"bilibili-recording-go/decode"
	"bilibili-recording-go/live"

	"github.com/gin-gonic/gin"
	// "github.com/kataras/golog"
)


type receiveInfo2 struct {
	RoomID string   `json:"roomID"`
}


func ProcessDecode(c *gin.Context) {
	info := receiveInfo2{}
	if c.ShouldBind(&info) == nil {
		fmt.Println(info)
		if _, ok := live.Lives[info.RoomID]; ok {
			l := live.CreateLiveSnapShot(live.Lives[info.RoomID])
			decode.SmartDecode(l)
			c.JSON(200, &struct {
				Msg bool `json:"msg"`
			}{
				true,
			})
			// if room.State != 5 {
			// 	golog.Debug(fmt.Sprintf("%s[RoomID: %s] 开始转码", room.Uname, info.RoomID))
			// 	room.DecodeStartTime = time.Now().Unix()
			// 	// room.Decode()
			// 	live.ManualDecode(info.RoomID)
			// 	golog.Debug(fmt.Sprintf("%s[RoomID: %s] 结束转码", room.Uname, info.RoomID))
			// 	room.DecodeEndTime = time.Now().Unix()
			// 	c.JSON(200, &struct {
			// 		Msg bool `json:"msg"`
			// 	}{
			// 		true,
			// 	})
			// } else {
			// 	c.JSON(200, &struct {
			// 		Msg bool `json:"msg"`
			// 	}{
			// 		false,
			// 	})
			// }
		} else {
			c.JSON(200, &struct {
				Msg bool `json:"msg"`
			}{
				false,
			})
		}
	}
}