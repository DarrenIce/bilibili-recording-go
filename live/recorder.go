package live

import (
	// "crypto/rand"
	"fmt"
	// "math/big"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"

	"bilibili-recording-go/config"
	"bilibili-recording-go/infos"
	"bilibili-recording-go/tools"
)

// GetInfoByRoom 获取Room info
func (r *Live) GetInfoByRoom(roomID string) {
	// n, _ := rand.Int(rand.Reader, big.NewInt(4))
	// time.Sleep((time.Duration(n.Int64()) + 1) * time.Second)
	// time.Sleep(1000 * time.Microsecond)
	defer func() {
		if v := recover(); v != nil {
			golog.Error("捕获了一个恐慌: ", v)
			return
		}
	}()
	url := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%s", roomID)
	req := requests.Requests()
	req.Proxy("socks5://127.0.0.1:1080")
	headers := requests.Header{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66",
		// "accept":	"application/json, text/javascript, */*; q=0.01",
		// "accept-encoding":	"gzip, deflate, br",
		// "accept-language":	"zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6,zh-TW;q=0.5",
		// "referer":	"https://live.bilibili.com/",
	}
	infs := infos.New()
	// req.Cookies = infs.BiliInfo.Cookies
	resp, err := req.Get(url, headers)
	if err != nil {
		golog.Error(err)
		return
	}
	data := gjson.Get(resp.Text(), "data")
	if resp.R.StatusCode == 200 {
		infs.UpdateFromGJSON(roomID, data)
		// golog.Debug("Get Room Info ", roomID)
	} else {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " 412啦, 快换代理")
	}
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
	// url, err := r.GetLiveURL(roomID)
	// if err != nil {
	// 	golog.Error(err)
	// }
	infs := infos.New()
	uname := infs.RoomInfos[roomID].Uname
	tools.Mkdir(fmt.Sprintf("./recording/%s/tmp", uname))
	outputName := uname + "_" + fmt.Sprint(time.Now().Format("20060102150405")) + ".flv"
	middle, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp", uname))
	outputFile := fmt.Sprint(middle + "\\" + outputName)
	url := fmt.Sprint("https://live.bilibili.com/", roomID)
	r.downloadCmds[roomID] = exec.Command("streamlink", "-f", "-o", outputFile, url, "best")
	// stdout, _ := r.downloadCmd.StdoutPipe()
	// r.downloadCmd.Stderr = r.downloadCmd.Stdout
	if err := r.downloadCmds[roomID].Start(); err != nil {
		golog.Error(err)
		r.downloadCmds[roomID].Process.Kill()
	}
	// tools.LiveOutput(stdout)
	r.downloadCmds[roomID].Wait()
	infs.RoomInfos[roomID].RecordEndTime = time.Now().Unix()
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 录制结束", infs.RoomInfos[roomID].Uname, roomID))
	r.unliveChannel <- roomID
}

func (r *Live) run(roomID string) {
	c := config.New()
	infs := infos.New()
	for {
		infs.UpadteFromConfig(roomID, c.Conf.Live[roomID])
		infs.RoomInfos[roomID].St, infs.RoomInfos[roomID].Et = tools.MkDuration(r.rooms[roomID].StartTime, r.rooms[roomID].EndTime)
		select {
		case rid := <-r.stop:
			if rid == roomID {
				if st, ok := r.syncMapGetUint32(roomID); ok && (st == running) {
					r.downloadCmds[roomID].Process.Kill()
				}
				infs.DeleteRoomInfo(roomID)
				return
			}
			r.stop <- rid
		default:
			if st, ok := r.syncMapGetUint32(roomID); ok && st == running && tools.JudgeInDuration(tools.MkDuration(r.rooms[roomID].StartTime, r.rooms[roomID].EndTime)) {
				time.Sleep(5 * time.Second)
			} else if r.judgeLive(roomID) && tools.JudgeInDuration(tools.MkDuration(r.rooms[roomID].StartTime, r.rooms[roomID].EndTime)) && infs.RoomInfos[roomID].AutoRecord {
				if st, ok := r.syncMapGetUint32(roomID); ok && (st == start || st == restart) {
					infs.RoomInfos[roomID].RecordStartTime = time.Now().Unix()
					infs.RoomInfos[roomID].RecordStatus = 1
					golog.Info(fmt.Sprintf("%s[RoomID: %s] 开始录制", infs.RoomInfos[roomID].Uname, roomID))
					go r.DownloadLive(roomID)
					if st == start {
						r.CompareAndSwapUint32(roomID, start, running)
					} else if st == restart {
						r.CompareAndSwapUint32(roomID, restart, running)
					}
				} else {
					time.Sleep(5 * time.Second)
				}
			} else if !tools.JudgeInDuration(tools.MkDuration(r.rooms[roomID].StartTime, r.rooms[roomID].EndTime)) {
				if st, ok := r.syncMapGetUint32(roomID); ok && st == restart {
					r.unliveChannel <- roomID
				} else if st, ok := r.syncMapGetUint32(roomID); ok && st == running {
					r.downloadCmds[roomID].Process.Kill()
					r.unliveChannel <- roomID
				} else {
					time.Sleep(5 * time.Second)
				}
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (r *Live) judgeLive(roomID string) bool {
	infs := infos.New()
	liveStatus := infs.RoomInfos[roomID].LiveStatus
	if liveStatus != 1 {
		return false
	}
	return true
}

func (r *Live) flushLiveStatus() {
	// delay := make(map[string]int)
	for {
		infs := infos.New()
		for roomID := range infs.RoomInfos {
			// if _, ok := delay[roomID]; !ok {
			r.GetInfoByRoom(roomID)
			time.Sleep(3 * time.Second)
			// }
		}
		// time.Sleep(1 * time.Second)
		// for k := range delay {
		// 	delay[k]--
		// }
		// for _, v := range infs.RoomInfos {
		// 	if _, ok := delay[v.RoomID]; v.LiveStatus != 1 && !ok {
		// 		delay[v.RoomID] = 5
		// 	}
		// }
		// deleteLst := []string{}
		// for k, v := range delay {
		// 	if v <=0 {
		// 		deleteLst = append(deleteLst, k)
		// 	}
		// }
		// for _, v := range deleteLst {
		// 	delete(delay, v)
		// }
	}
}

func (r *Live) unlive() {
	for {
		select {
		case roomID := <-r.unliveChannel:
			if tools.JudgeInDuration(tools.MkDuration(r.rooms[roomID].StartTime, r.rooms[roomID].EndTime)) {
				time.Sleep(10 * time.Second)
				r.CompareAndSwapUint32(roomID, running, restart)
			} else {
				if r.CompareAndSwapUint32(roomID, running, waiting) || r.CompareAndSwapUint32(roomID, restart, waiting) {
					r.decodeChannel <- roomID
				}
			}
		}
	}
}

func (r *Live) recordWorker() {
	go r.unlive()
	for {
		info := <-r.recordChannel
		roomID := info.RoomID
		r.rooms[roomID] = info
		golog.Info(fmt.Sprintf("房间[RoomID: %s] 开始监听", roomID))
		go r.start(roomID)
	}
}

func (r *Live) start(roomID string) {
	r.CompareAndSwapUint32(roomID, iinit, start)
	go r.run(roomID)
}

// Stop 现在是所有状态都可以转移到stop，会有点问题，如果在转码或者上传期间stop，会有roomID取不到值的报错，可以考虑加一个协程判断state，state回归后再delete
func (r *Live) Stop(roomID string) {
	golog.Info(fmt.Sprintf("房间[RoomID: %s] 退出监听", roomID))
	infs := infos.New()
	r.state.Store(roomID, stop)
	s, _ := r.state.Load(roomID)
	infs.RoomInfos[roomID].State, _ = s.(uint32)
	r.stop <- roomID
}


