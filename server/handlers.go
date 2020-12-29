package server

import (
	"net/http"

	"bilibili-recording-go/config"
	"bilibili-recording-go/tools"
	"bilibili-recording-go/infos"
)

func getAllLives(writer http.ResponseWriter, r *http.Request) {
	lives := []string{}
	config, _ := config.LoadConfig()
	for _, v := range config.Live {
		lives = append(lives, v.RoomID)
	}
	tools.WriteJSON(writer, lives)
}

func getAllInfos(writer http.ResponseWriter, r *http.Request) {
	infs := infos.New()
	tools.WriteJSON(writer, infs)
}