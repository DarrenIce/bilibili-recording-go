package live

import (
	"fmt"
	"os"
	"strconv"
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
			if r.judgeRecord() && r.judgeLive() && r.judegArea() {
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
			} else if !r.judgeRecord() || !r.judegArea() {
				if r.State == restart {
					r.unlive()
				} else if r.State == running {
					r.downloadCmd.Process.Kill()
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
	return r.LiveStatus == 1
}

func (r *Live) unlive() {
	// if r.judgeRecord() && r.judegArea() && !r.RecordMode {
	// 	time.Sleep(10 * time.Second)
	// 	atomic.CompareAndSwapUint32(&r.State, running, restart)
	// } else {
	if atomic.CompareAndSwapUint32(&r.State, running, waiting) || atomic.CompareAndSwapUint32(&r.State, restart, waiting) {
		if r.SaveDanmu && r.Platform == "bilibili" {
			r.danmuClient.Stop <- struct{}{}
		}
		if tools.GetTimeDeltaFromTimestamp(r.RecordEndTime, r.RecordStartTime) < 60 {
			atomic.CompareAndSwapUint32(&r.State, waiting, start)
			if r.SaveDanmu && r.Platform == "bilibili" {
				os.Remove(fmt.Sprintf("./recording/%s/%s.ass", r.Uname, r.danmuClient.Ass.File))
			}
			return
		}
		DecodeChan <- CreateLiveSnapShot(r)
		time.Sleep(5 * time.Second)
		atomic.CompareAndSwapUint32(&r.State, waiting, start)
	}
	// }
}

func (r *Live) start() {
	golog.Info(fmt.Sprintf("房间[RoomID: %s] 开始监听", r.RoomID))
	atomic.CompareAndSwapUint32(&r.State, iinit, start)
	r.UpdateSiteInfo()
	go r.run()
}

func (r *Live) judegArea() bool {
	if !r.AreaLock {
		return true
	}
	return r.AreaLimit == r.AreaName
}

func (r *Live) judgeRecord() bool {
	return r.AutoRecord && (r.RecordMode || (!r.RecordMode && tools.JudgeInDuration(r.St, r.Et)))
}
