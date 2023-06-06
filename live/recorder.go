package live

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kataras/golog"

	"bilibili-recording-go/danmu"
	"bilibili-recording-go/tools"
)



func (r *Live) run() {
	for {
		select {
		case <-r.stop:
			if r.State == running {
				r.downloadCmd.Process.Kill()
			}
			return
		default:
			r.St, r.Et = tools.MkDuration(r.StartTime, r.EndTime)
			if r.judgeRecord() && r.judgeLive() && r.judgeArea() {
				if r.State == running {
					time.Sleep(5 * time.Second)
				} else if r.State == start || r.State == restart {
					r.RecordStartTime = time.Now().Unix()
					r.RecordStatus = 1
					golog.Info(fmt.Sprintf("%s[RoomID: %s] 开始录制", r.Uname, r.RoomID))
					go r.site.DownloadLive(r)
					if r.SaveDanmu && r.Platform == "bilibili" {
						file := fmt.Sprintf("%s_%s场_%s_%s", r.Uname, time.Unix(r.RecordStartTime, 0).Format("2006-01-02 15时04分"), r.AreaName, r.Title)
						roomID_uint64, _ := strconv.ParseUint(r.RoomID, 10, 64)
						r.danmuClient = danmu.NewDanmuClient(uint32(roomID_uint64), r.Uname, file)
						go r.danmuClient.Run()
					}
					if r.State == start {
						atomic.CompareAndSwapUint32(&r.State, start, running)
					} else if r.State == restart {
						atomic.CompareAndSwapUint32(&r.State, restart, running)
					}
				} else {
					time.Sleep(5 * time.Second)
				}
			} else {
				if r.State == restart {
					r.unlive()
				} else if r.State == running && r.downloadCmd.Process != nil && (!r.judgeRecord() || !r.judgeArea()) {
					r.downloadCmd.Process.Kill()
				} else {
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}

func (r *Live) judgeLive() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.LiveStatus == 1
}

func (r *Live) unlive() {
	// if r.judgeRecord() && r.judgeArea() && !r.RecordMode {
	// 	time.Sleep(10 * time.Second)
	// 	atomic.CompareAndSwapUint32(&r.State, running, restart)
	// } else {
	if atomic.CompareAndSwapUint32(&r.State, running, waiting) || atomic.CompareAndSwapUint32(&r.State, restart, waiting) {
		for r.SaveDanmu && r.Platform == "bilibili" && !r.danmuClient.AfterConnected {
			time.Sleep(1 * time.Second)
		}
		if r.SaveDanmu && r.Platform == "bilibili" && r.danmuClient.Connected {
			r.danmuClient.Stop <- struct{}{}
		}
		if tools.GetTimeDeltaFromTimestamp(r.RecordEndTime, r.RecordStartTime) < 60 {
			time.Sleep(120 * time.Second)
			atomic.CompareAndSwapUint32(&r.State, waiting, start)
			if r.SaveDanmu && r.Platform == "bilibili" {
				os.Remove(fmt.Sprintf("./recording/%s/ass/%s.ass", r.Uname, r.danmuClient.Ass.File))
				os.Remove(fmt.Sprintf("./recording/%s/brg/%s.brg", r.Uname, r.danmuClient.Brg.File))
			}
			return
		}
		DecodeChan <- CreateLiveSnapShot(r)
		r.TmpFilePath = ""
		time.Sleep(5 * time.Second)
		atomic.CompareAndSwapUint32(&r.State, waiting, start)
	}
	// }
}

func (r *Live) start() {
	golog.Info(fmt.Sprintf("房间[RoomID: %s] 开始监听", r.RoomID))
	atomic.CompareAndSwapUint32(&r.State, iinit, start)
	if r.Uname != "" {
		tools.Mkdir(fmt.Sprintf("./recording/%s/brg", r.Uname))
		tools.Mkdir(fmt.Sprintf("./recording/%s/tmp", r.Uname))
		tools.Mkdir(fmt.Sprintf("./recording/%s/ass", r.Uname))
	}
	r.UpdateSiteInfo()
	go r.run()
}

func (r *Live) judgeArea() bool {
	if !r.AreaLock {
		return true
	}
	for _, v := range strings.Split(r.AreaLimit, ",") {
		if v == r.AreaName {
			return true
		}
	}
	return false
}

func (r *Live) judgeRecord() bool {
	return r.AutoRecord && (r.RecordMode || (!r.RecordMode && tools.JudgeInDuration(r.St, r.Et)))
}
