package server

import (
	"io/ioutil"
	"net/http"

	"bilibili-recording-go/config"
	"bilibili-recording-go/infos"
	"bilibili-recording-go/live"
	"bilibili-recording-go/tools"

	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
)

func getAllLives(writer http.ResponseWriter, r *http.Request) {
	lives := []string{}
	c := config.New()
	for _, v := range c.Conf.Live {
		lives = append(lives, v.RoomID)
	}
	tools.WriteJSON(writer, lives)
	// golog.Info("getAllLives Success!")
}

func getAllInfos(writer http.ResponseWriter, r *http.Request) {
	infs := infos.New()
	tools.WriteJSON(writer, infs)
	// golog.Info("getAllInfos Success!")
}

func saveConfig(writer http.ResponseWriter, r *http.Request) {
	c := config.New()
	err := c.Marshal()
	status := make(map[string]string)
	if err != nil {
		golog.Error(err)
		status["info"] = "saveConfig Error"
	} else {
		golog.Info("saveConfig Success!")
		status["info"] = "saveConfig Success"
	}
	tools.WriteJSON(writer, status)
}

func addRooms(writer http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	resps := []live.InfoResponse{}
	liver := live.New()
	c := config.New()
	gjson.ParseBytes(b).ForEach(func(key, value gjson.Result) bool {
		roomConfigInfo := config.RoomConfigInfo{}
		roomConfigInfo.RoomID = value.Get("RoomID").String()
		roomConfigInfo.StartTime = value.Get("StartTime").String()
		roomConfigInfo.EndTime = value.Get("EndTime").String()
		roomConfigInfo.AutoRecord = value.Get("AutoRecord").Bool()
		roomConfigInfo.AutoUpload = value.Get("AutoUpload").Bool()
		c.AddRoom(roomConfigInfo.RoomID, roomConfigInfo)
		liver.AddRoom(roomConfigInfo)
		resp, err := live.GetRoomInfoForResp(roomConfigInfo)
		if err != nil {
			return false
		}
		resps = append(resps, resp)
		return true
	})
	c.Marshal()
	tools.WriteJSON(writer, resps)
}

func deleteRooms(writer http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	c := config.New()
	liver := live.New()
	gjson.ParseBytes(b).ForEach(func(key, value gjson.Result) bool {
		c.DeleteRoom(value.String())
		liver.DeleteRoom(value.String())
		return true
	})
	c.Marshal()
	status := make(map[string]string)
	golog.Info("deleteRooms Success!")
	status["info"] = "deleteRooms Success"
	tools.WriteJSON(writer, status)
}