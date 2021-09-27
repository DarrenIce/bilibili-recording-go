package server

import (
	"io/ioutil"
	"net/http"

	"bilibili-recording-go/config"
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
	tools.WriteJSON(writer, live.Lives)
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
	c := config.New()
	gjson.ParseBytes(b).ForEach(func(key, value gjson.Result) bool {
		roomConfigInfo := config.RoomConfigInfo{}
		roomConfigInfo.RoomID = value.Get("RoomID").String()
		roomConfigInfo.StartTime = value.Get("StartTime").String()
		roomConfigInfo.EndTime = value.Get("EndTime").String()
		roomConfigInfo.AutoRecord = value.Get("AutoRecord").Bool()
		roomConfigInfo.AutoUpload = value.Get("AutoUpload").Bool()
		c.AddRoom(roomConfigInfo)
		live.AddRoom(roomConfigInfo.RoomID)
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
	gjson.ParseBytes(b).ForEach(func(key, value gjson.Result) bool {
		c.DeleteRoom(value.String())
		live.DeleteRoom(value.String())
		return true
	})
	c.Marshal()
	status := make(map[string]string)
	golog.Info("deleteRooms Success!")
	status["info"] = "deleteRooms Success"
	tools.WriteJSON(writer, status)
}

func manualDecode(writer http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	gjson.ParseBytes(b).ForEach(func(key, value gjson.Result) bool {
		if live.ManualDecode(value.String()) {
			return true
		}
		return false
	})
	status := make(map[string]string)
	golog.Info("manualDecode Success!")
	status["info"] = "manualDecode Success"
	tools.WriteJSON(writer, status)
}

func manualUpload(writer http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	gjson.ParseBytes(b).ForEach(func(key, value gjson.Result) bool {
		if live.ManualUpload(value.String()) {
			return true
		}
		return false
	})
	status := make(map[string]string)
	golog.Info("manualUpload Success!")
	status["info"] = "manualUpload Success"
	tools.WriteJSON(writer, status)
}