package monitor

import (
	"bilibili-recording-go/config"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	// "github.com/kataras/golog"
	"github.com/tidwall/gjson"
)

type MonitorRoom struct {
	RoomID     string
	UID        string
	Uname      string
	Title      string
	Popularity int64
	ParentID   string
	ParentName string
	AreaID     string
	AreaName   string
	UserCover  string
	LiveCover  string
}

type MonitorRoomSlice []MonitorRoom

func (mrs MonitorRoomSlice) Len() int {
	return len(mrs)
}

func (mrs MonitorRoomSlice) Less(i, j int) bool {
	return mrs[i].Popularity > mrs[j].Popularity
}

func (mrs MonitorRoomSlice) Swap(i, j int) {
	mrs[i], mrs[j] = mrs[j], mrs[i]
}

type AreaList struct {
	Data MonitorRoomSlice
	Nums int
}

var (
	Lock           *sync.Mutex
	// AreaMonitorMap = make(map[string]AreaList)
	AreaInfoList 	MonitorRoomSlice
	areaMonitorApi = "https://api.live.bilibili.com/xlive/web-interface/v1/second/getList?platform=web&parent_area_id=%s&area_id=%s&sort_type=&page=%d"
)

func init() {
	Lock = new(sync.Mutex)
	go Monitor()
}

func Monitor() {
	for {
		areaInfoList := make(MonitorRoomSlice, 0)
		c := config.New()
		areas := make([]config.MonitorArea, 0)
		for _, v := range c.Conf.MonitorAreas {
			if v.AreaName == "" {
				continue
			}
			tmparea := config.MonitorArea{
				AreaID:   v.AreaID,
				AreaName: v.AreaName,
				ParentID: v.ParentID,
				Platform: v.Platform,
			}
			areas = append(areas, tmparea)
		}
		for _, v := range areas {
			page := 1
			// areaname := v.AreaName
			// areaList := &AreaList{
			// 	Data: make([]MonitorRoom, 0),
			// 	Nums: 0,
			// }
			uidmap := make(map[string]string)
			for {
				url := fmt.Sprintf(areaMonitorApi, v.ParentID, v.AreaID, page)
				resp, err := requests.Get(url)
				if err != nil {
					// golog.Error(err)
					time.Sleep(10 * time.Minute)
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
					userCover := room.Get("user_cover").String()
					liveCover := room.Get("cover").String()
					if judgeRoomBlocked(roomID) {
						continue
					}
					if err != nil {
						golog.Error(err)
					}
					if _, ok := uidmap[roomID]; !ok {
						uidmap[roomID] = uid
						Lock.Lock()
						areaInfoList = append(areaInfoList, MonitorRoom{
							RoomID:     roomID,
							UID:        uid,
							Uname:      uname,
							Title:      title,
							Popularity: popularity,
							ParentID:   parentID,
							ParentName: parentName,
							AreaID:     areaID,
							AreaName:   areaName,
							UserCover:  userCover,
							LiveCover:  liveCover,
						})
						Lock.Unlock()
						// areaList.Nums++
					}
				}
				page++
				time.Sleep(time.Second * 3)
			}
			Lock.Lock()
			sort.Sort(areaInfoList)
			Lock.Unlock()
			// AreaMonitorMap[areaname] = *areaList
			time.Sleep(time.Second * 10)
		}
		AreaInfoList = areaInfoList
		time.Sleep(30 * time.Second)
	}
}

func judgeRoomBlocked(roomID string) bool {
	c := config.New()
	for _, v := range c.Conf.BlockedRooms {
		if v == roomID {
			return true
		}
	}
	return false
}