package live

import (
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"

	"bilibili-recording-go/config"
	"bilibili-recording-go/tools"
)

func Upload2BaiduPCS() {
	golog.Info("Upload2BaiduPCS Start")
	for _, v := range Lives {
		uname := v.Uname
		localBasePath := fmt.Sprint("./recording/", uname)
		if !tools.Exists(localBasePath) {
			continue
		}
		pcsBasePath := fmt.Sprint("/录播/", uname)
		cmd := exec.Command("./BaiduPCS-Go.exe", "mkdir", pcsBasePath)
		stdout, _ := cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		cmd.Start()
		tools.LiveOutput(stdout)
		for _, f := range tools.ListDir(localBasePath) {
			if o, _ := os.Stat(f); !o.IsDir() {
				cmd = exec.Command("./BaiduPCS-Go.exe", "upload", f, pcsBasePath)
				stdout, _ := cmd.StdoutPipe()
				cmd.Stderr = cmd.Stdout
				cmd.Start()
				tools.LiveOutput(stdout)
			}
		}
	}
	golog.Info("Upload2BaiduPCS End")
}

// AddRoom ADD
func AddRoom(roomID string) {
	live := new(Live)
	live.Init((roomID))
	LmapLock.Lock()
	Lives[roomID] = live
	LmapLock.Unlock()
	go live.start()
}

// DeleteRoom deleteroom
func DeleteRoom(roomID string) {
	LmapLock.Lock()
	defer LmapLock.Unlock()
	live := new(Live)
	if _, ok := Lives[roomID]; ok {
		live = Lives[roomID]
		delete(Lives, roomID)
	}
	live.stop <- struct{}{}
	// 如何释放？
	// State == start || State == restart
}

// GetRoomInfoForResp get
func GetRoomInfoForResp(info config.RoomConfigInfo) (InfoResponse, error) {
	url := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%s", info.RoomID)
	resp, err := requests.Get(url)
	if err != nil {
		return InfoResponse{}, err
	}
	data := gjson.Get(resp.Text(), "data")
	inf := InfoResponse{}
	inf.RoomID = info.RoomID
	inf.AutoRecord = info.AutoRecord
	inf.AutoUpload = info.AutoUpload
	inf.RealID = data.Get("room_info").Get("room_id").String()
	inf.LiveStatus = int(data.Get("room_info").Get("live_status").Int())
	inf.Uname = data.Get("anchor_info").Get("base_info").Get("uname").String()
	inf.Title = data.Get("room_info").Get("title").String()
	return inf, nil
}

func ManualUpload(roomID string) bool {
	LmapLock.Lock()
	defer LmapLock.Unlock()
	if _, ok := Lives[roomID]; ok {
		if atomic.CompareAndSwapUint32(&Lives[roomID].State, start, uploadWait) || Lives[roomID].State == uploadWait {
			uploadChan <- roomID
			return true
		}
	}
	return false
}

func ManualDecode(roomID string) bool {
	LmapLock.Lock()
	defer LmapLock.Unlock()
	if _, ok := Lives[roomID]; ok {
		if atomic.CompareAndSwapUint32(&Lives[roomID].State, start, waiting) {
			decodeChan <- roomID
			return true
		}
	}
	return false
}

func flushLiveStatus() {
	for {
		lst := make([]string, len(Lives))
		LmapLock.Lock()
		for k := range Lives {
			lst = append(lst, k)
		}
		LmapLock.Unlock()
		for _, v := range lst {
			LmapLock.Lock()
			live, ok := Lives[v]
			if !ok {
				LmapLock.Unlock()
				continue
			}
			LmapLock.Unlock()
			live.UpdateSiteInfo()
			time.Sleep(10 * time.Second)
		}
	}
}