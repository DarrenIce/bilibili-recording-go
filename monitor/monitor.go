package monitor

import (
	"fmt"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
)

type MonitorRoom struct {
	RoomID string
	UID	string
	Uname string
	Title string
	Popularity int64
	ParentID string
	ParentName string
	AreaID	string
	AreaName string
}

type AreaList []*MonitorRoom

var (
	MonitorMap = make(map[string]AreaList)
	ParentIDs = []int{5}
	AreaIDs = []int{339}
	monitorApi = "https://api.live.bilibili.com/xlive/web-interface/v1/second/getList?platform=web&parent_area_id=%d&area_id=%d&sort_type=&page=%d"
)

func Monitor() {
	for {
		for k := range ParentIDs {
			page := 1
			for {
				url := fmt.Sprintf(monitorApi, ParentIDs[k], AreaIDs[k], page)
				resp, err := requests.Get(url)
				if err != nil {
					golog.Error(err)
					continue
				}
				if resp.R.StatusCode != 200 {
					fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " 412啦, 快换代理")
					time.Sleep(1 * time.Hour)
					continue
				}
				if gjson.Get(resp.Text(), "code").Int() != 0 {
					continue
				}
				data := gjson.Get(resp.Text(), "data")
				if len(data.Get("list").Array()) == 0 {
					break
				}
				for _, room := range data.Get("list").Array() {
					roomID := room.Get("roomid").String()
					uid := room.Get("uid").String()
					uname := room.Get("uname").String()
					title := room.Get("title").String()
					popularity := room.Get("online").Int()
					parentID := room.Get("parent_id").String()
					parentName := room.Get("parent_name").String()
					areaID := room.Get("area_id").String()
					areaName := room.Get("area_name").String()
					MonitorMap[roomID] = append(MonitorMap[roomID], &MonitorRoom{
						RoomID: roomID,
						UID: uid,
						Uname: uname,
						Title: title,
						Popularity: popularity,
						ParentID: parentID,
						ParentName: parentName,
						AreaID: areaID,
						AreaName: areaName,
					})
				}
				page++
				time.Sleep(time.Second * 3)
			}
			time.Sleep(time.Second * 30)
		}
		time.Sleep(5 * time.Minute)
	}
}