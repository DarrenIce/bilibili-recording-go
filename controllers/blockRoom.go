package controllers

import (
	"encoding/json"
	"fmt"

	"bilibili-recording-go/config"
	"bilibili-recording-go/monitor"
	"bilibili-recording-go/tools"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/kataras/golog"
)

type BlockRoomController struct {
	beego.Controller
}

func (c *BlockRoomController) Post() {
	fmt.Println(tools.BytesToStringFast(c.Ctx.Input.RequestBody))
	info := new(receiveInfo2)
	json.Unmarshal(c.Ctx.Input.RequestBody, info)
	fmt.Println(info)
	c.Data["json"] = &struct {
		Msg bool `json:"msg"`
	}{
		addBlockRoom(info.RoomID),
	}
	c.ServeJSON()
}

func addBlockRoom(roomID string) bool {
	c := config.New()
	for _, v := range c.Conf.BlockedRooms {
		if v == roomID {
			return false
		}
	}
	c.AddBlockedRoom(roomID)
	if err := c.Marshal(); err != nil {
		golog.Error(err)
		return false
	}
	monitor.Lock.Lock()
	defer monitor.Lock.Unlock()
	for k := range monitor.AreaMonitorMap {
		for kk, v := range monitor.AreaMonitorMap[k].Data {
			if v.RoomID == roomID {
				newData := append(monitor.AreaMonitorMap[k].Data[:kk], monitor.AreaMonitorMap[k].Data[kk+1:]...)
				numNum := monitor.AreaMonitorMap[k].Nums - 1
				newArea := &monitor.AreaList{
					Data: newData,
					Nums: numNum,
				}
				monitor.AreaMonitorMap[k] = *newArea
			}
		}
	}
	return true
}
