package live

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"

	"bilibili-recording-go/config"
	"bilibili-recording-go/infos"
	"bilibili-recording-go/tools"
)

// GetInfoByRoom 获取Room info
func (r *Live) GetInfoByRoom(roomID string) error {
	url := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%s", roomID)
	resp, err := requests.Get(url)
	if err != nil {
		return err
	}
	data := gjson.Get(resp.Text(), "data")
	infs := infos.New()
	infs.UpdateFromGJSON(roomID, data)
	return nil
}

// GetLiveURL 获取Room live url
func (r *Live) GetLiveURL(roomID string) (string, error) {
	url := "https://api.live.bilibili.com/xlive/web-room/v1/playUrl/playUrl"
	infs := infos.New()
	paras := requests.Params{
		"cid":      infs.RoomInfos[roomID].RealID,
		"qn":       "4",
		"platform": "web",
	}
	resp, err := requests.Get(url, paras)
	if err != nil {
		return "", err
	}
	liveURL := gjson.Get(resp.Text(), "data.durl.0.url").String()
	return liveURL, nil
}

// DownloadLive 下载
func (r *Live) DownloadLive(roomID string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	url, err := r.GetLiveURL(roomID)
	if err != nil {
		golog.Error(err)
	}
	infs := infos.New()
	uname := infs.RoomInfos[roomID].Uname
	tools.Mkdir(fmt.Sprintf("./recording/%s/tmp", uname))
	outputName := uname + "_" + fmt.Sprint(time.Now().Format("20060102150405")) + ".flv"
	r.downloadCmd = exec.Command("ffmpeg", "-i", url, "-c", "copy", fmt.Sprintf("./recording/%s/tmp/%s", uname, outputName))
	// stdout, _ := r.downloadCmd.StdoutPipe()
	// r.downloadCmd.Stderr = r.downloadCmd.Stdout
	if err = r.downloadCmd.Start(); err != nil {
		r.downloadCmd.Process.Kill()
		golog.Fatal(err)
	}
	// tools.LiveOutput(stdout)
	r.downloadCmd.Wait()
	infs.RoomInfos[roomID].RecordEndTime = time.Now().Format("2006-01-02 15:04:05")
	golog.Debug(fmt.Sprintf("%s[RoomID: %s] 录制结束", infs.RoomInfos[roomID].Uname, roomID))
	r.unliveChannel <- roomID
}

func (r *Live) run(roomID string) {
	config, _ := config.LoadConfig()
	infs := infos.New()
	for {
		infs.UpadteFromConfig(roomID, config.Live[roomID])
		infs.RoomInfos[roomID].St, infs.RoomInfos[roomID].Et = tools.MkDuration(r.rooms[roomID].StartTime, r.rooms[roomID].EndTime)
		select {
		case <-r.stop:
			r.downloadCmd.Process.Kill()
			return
		default:
			if r.judgeLive(roomID) && tools.JudgeInDuration(tools.MkDuration(r.rooms[roomID].StartTime, r.rooms[roomID].EndTime)) && infs.RoomInfos[roomID].AutoRecord {
				if st, ok := r.syncMapGetUint32(roomID); ok && st == start {
					infs.RoomInfos[roomID].RecordStartTime = time.Now().Format("2006-01-02 15:04:05")
					infs.RoomInfos[roomID].RecordStatus = 1
					golog.Debug(fmt.Sprintf("%s[RoomID: %s] 开始录制", infs.RoomInfos[roomID].Uname, roomID))
					go r.DownloadLive(roomID)
					r.compareAndSwapUint32(roomID, start, running)
				}
			} else {
				time.Sleep(3 * time.Second)
			}
		}
	}
}

func (r *Live) judgeLive(roomID string) bool {
	err := r.GetInfoByRoom(roomID)
	if err != nil {
		golog.Error(err)
	}
	infs := infos.New()
	liveStatus := infs.RoomInfos[roomID].LiveStatus
	if liveStatus != 1 {
		return false
	}
	return true
}

func (r *Live) unlive() {
	for {
		select {
		case roomID := <-r.unliveChannel:
			if tools.JudgeInDuration(tools.MkDuration(r.rooms[roomID].StartTime, r.rooms[roomID].EndTime)) {
				r.compareAndSwapUint32(roomID, running, start)
			} else {
				r.compareAndSwapUint32(roomID, running, waiting)
				r.decodeChannel <- roomID
			}
		}
	}
}

func (r *Live) recordWorker() {
	for {
		info := <-r.recordChannel
		roomID := info.RoomID
		r.rooms[roomID] = info
		golog.Debug(fmt.Sprintf("房间[RoomID: %s] 开始监听", roomID))
		go r.start(roomID)
	}
}

func (r *Live) start(roomID string) {
	r.compareAndSwapUint32(roomID, iinit, start)
	go r.run(roomID)
	go r.unlive()
}

// Stop stop
func (r *Live) Stop(roomID string) {
	infs := infos.New()
	r.state.Store(roomID, stop)
	s, _ := r.state.Load(roomID)
	infs.RoomInfos[roomID].State, _ = s.(uint32)
	close(r.stop)
}

func (r *Live) compareAndSwapUint32(roomID string, old uint32, new uint32) bool {
	s, _ := r.state.Load(roomID)
	st, _ := s.(uint32)
	if st == old {
		r.state.Store(roomID, new)
		infs := infos.New()
		infs.RoomInfos[roomID].State = new
		roomInfo := infs.RoomInfos[roomID]
		golog.Debug(fmt.Sprintf("%s[RoomID: %s] state changed from %d to %d", roomInfo.Uname, roomInfo.RoomID, old, new))
		return true
	}
	return false
}

func (r *Live) syncMapGetUint32(roomID string) (uint32, bool) {
	s, ok := r.state.Load(roomID)
	if !ok {
		return 0, ok
	}
	st, ok := s.(uint32)
	if !ok {
		return 0, ok
	}
	return st, true
}
