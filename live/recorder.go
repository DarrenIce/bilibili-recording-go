package live

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
	"sync/atomic"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"

	"bilibili-recording-go/config"
	"bilibili-recording-go/tools"
)

// GetInfoByRoom 获取Room info(goroutine)
func (r *Live) GetInfoByRoom() {
	defer func() {
		if v := recover(); v != nil {
			golog.Error("捕获了一个恐慌: ", v)
			return
		}
	}()
	url := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%s", r.RoomID)
	req := requests.Requests()
	c := config.New()
	if c.Conf.RcConfig.NeedProxy {
		req.Proxy(c.Conf.RcConfig.Proxy)
	}
	headers := requests.Header{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66",
		// "accept":	"application/json, text/javascript, */*; q=0.01",
		// "accept-encoding":	"gzip, deflate, br",
		// "accept-language":	"zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6,zh-TW;q=0.5",
		// "referer":	"https://live.bilibili.com/",
	}
	// req.Cookies = infs.BiliInfo.Cookies
	resp, err := req.Get(url, headers)
	if err != nil {
		golog.Error(err)
		return
	}
	data := gjson.Get(resp.Text(), "data")
	if resp.R.StatusCode == 200 {
		r.UpdateFromGJSON(data)
	} else {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " 412啦, 快换代理")
	}
}

// GetLiveURL 获取Room live url(停用)
func (r *Live) GetLiveURL(roomID string) (string, error) {
	url := "https://api.live.bilibili.com/xlive/web-room/v1/playUrl/playUrl"
	paras := requests.Params{
		"cid":      Lives[roomID].RealID,
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
func (r *Live) DownloadLive() {
	// url, err := r.GetLiveURL(roomID)
	// if err != nil {
	// 	golog.Error(err)
	// }
	uname := r.Uname
	tools.Mkdir(fmt.Sprintf("./recording/%s/tmp", uname))
	outputName := uname + "_" + fmt.Sprint(time.Now().Format("20060102150405")) + ".flv"
	middle, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp", uname))
	outputFile := fmt.Sprint(middle + "\\" + outputName)
	url := fmt.Sprint("https://live.bilibili.com/", r.RoomID)
	r.downloadCmd = exec.Command("streamlink", "-f", "-o", outputFile, url, "best")
	// stdout, _ := r.downloadCmd.StdoutPipe()
	// r.downloadCmd.Stderr = r.downloadCmd.Stdout
	if err := r.downloadCmd.Start(); err != nil {
		golog.Error(err)
		r.downloadCmd.Process.Kill()
	}
	// tools.LiveOutput(stdout)
	r.downloadCmd.Wait()
	r.RecordEndTime = time.Now().Unix()
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 录制结束", r.Uname, r.RoomID))
	r.unlive()
}

func (r *Live) run() {
	for {
		select {
		case <-r.stop:
			if r.State == running {
				r.downloadCmd.Process.Kill()
			}
			return
		default:
			if r.State == running && tools.JudgeInDuration(r.St, r.Et) {
				time.Sleep(5 * time.Second)
			} else if r.judgeLive() && tools.JudgeInDuration(r.St, r.Et) && r.AutoRecord {
				if r.State == start || r.State == restart {
					r.RecordStartTime = time.Now().Unix()
					r.RecordStatus = 1
					golog.Info(fmt.Sprintf("%s[RoomID: %s] 开始录制", r.Uname, r.RoomID))
					go r.DownloadLive()
					if r.State == start {
						atomic.CompareAndSwapUint32(&r.State, start, running)
					} else if r.State == restart {
						atomic.CompareAndSwapUint32(&r.State, restart, running)
					}
				} else {
					time.Sleep(5 * time.Second)
				}
			} else if !tools.JudgeInDuration(r.St, r.Et) {
				if r.State == restart {
					r.unlive()
				} else if r.State == running {
					r.downloadCmd.Process.Kill()
					r.unlive()
				} else {
					time.Sleep(5 * time.Second)
				}
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (r *Live) judgeLive() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.LiveStatus != 1 {
		return false
	}
	return true
}

func (r *Live) unlive() {
	if tools.JudgeInDuration(r.St, r.Et) {
		time.Sleep(10 * time.Second)
		atomic.CompareAndSwapUint32(&r.State, running, restart)
	} else {
		if atomic.CompareAndSwapUint32(&r.State, running, waiting) || atomic.CompareAndSwapUint32(&r.State, restart, waiting) {
			decodeChan <- r.RoomID
		}
	}
}

func (r *Live) start() {
	golog.Info(fmt.Sprintf("房间[RoomID: %s] 开始监听", r.RoomID))
	atomic.CompareAndSwapUint32(&r.State, iinit, start)
	r.GetInfoByRoom()
	go r.run()
}
