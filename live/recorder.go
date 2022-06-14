package live

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/kataras/golog"

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
		if tools.GetTimeDeltaFromTimestamp(r.RecordEndTime, r.RecordStartTime) < 60 {
			atomic.CompareAndSwapUint32(&r.State, waiting, start)
			return
		}
		DecodeChan <- CreateLiveSnapShot(r)
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
