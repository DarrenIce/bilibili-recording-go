package live

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
	PRLock.Lock()
	PlatformRooms[live.Platform] = append(PlatformRooms[live.Platform], roomID)
	PRLock.Unlock()
	fmt.Printf("AddRoom %s, 当前队列: %s\n", roomID, PlatformRooms[live.Platform])
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
	PRLock.Lock()
	defer PRLock.Unlock()
	for i, v := range PlatformRooms[live.Platform] {
		if v == roomID {
			PlatformRooms[live.Platform] = append(PlatformRooms[live.Platform][:i], PlatformRooms[live.Platform][i+1:]...)
			break
		}
	}
	fmt.Printf("DeleteRoom %s, 当前队列: %s\n", roomID, PlatformRooms[live.Platform])
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

func flushPlatformLives(platform string) {
	// if platform == "douyin" {
	// 	fmt.Println("pause douyin support.")
	// 	return
	// }
	for {
		lst := make([]string, 0)
		PRLock.Lock()
		lst = append(lst, PlatformRooms[platform]...)
		PRLock.Unlock()
		if len(lst) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}
		for _, v := range lst {
			LmapLock.Lock()
			live, ok := Lives[v]
			if !ok {
				LmapLock.Unlock()
				continue
			}
			LmapLock.Unlock()
			live.UpdateSiteInfo()
			if (platform == "douyin") {
				time.Sleep(15 * time.Second)
			} else {
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func CreateLiveSnapShot(live *Live) *LiveSnapshot {
	fmt.Println("CreateLiveSnapShot")
	snapshot := LiveSnapshot{}
	snapshot.SiteInfo = live.SiteInfo
	snapshot.State = live.State
	snapshot.RoomConfigInfo = live.RoomConfigInfo
	if live.LiveStatus == 1 {
		snapshot.TmpFilePath = live.TmpFilePath
	} else {
		snapshot.TmpFilePath = ""
	}
	return &snapshot
}

func CleanRecordingDir() {
	golog.Info("CleanRecordingDir Start")
	for _, v := range Lives {
		if v.CleanUpRegular {
			tmpDir := fmt.Sprintf("./recording/%s/tmp", v.Uname)
			userDir := fmt.Sprintf("./recording/%s", v.Uname)
			timeNowStamp := time.Now().Unix()
			expireTime := tools.ConvertString2TimeStamp(v.SaveDuration)
			if tools.Exists(tmpDir) {
				for _, f := range tools.ListDir(tmpDir) {
					if ok := strings.HasSuffix(f, ".flv") || strings.HasSuffix(f, ".mts"); ok {
						fileModifyTime := tools.GetFileLastModifyTime(f)
						if tools.GetTimeDeltaFromTimestamp(timeNowStamp, fileModifyTime) > expireTime {
							err := os.Remove(f)
							if err != nil {
								golog.Error(err.Error())
							} else {
								golog.Info("Remove Expire Tmp File: ", f)
							}
						}
					}
				}
			}
			if tools.Exists(userDir) {
				for _, f := range tools.ListDir(userDir) {
					if ok := strings.HasSuffix(f, ".mp4") || strings.HasSuffix(f, ".m4a"); ok {
						fileModifyTime := tools.GetFileLastModifyTime(f)
						if tools.GetTimeDeltaFromTimestamp(timeNowStamp, fileModifyTime) > expireTime {
							err := os.Remove(f)
							if err != nil {
								golog.Error(err.Error())
							} else {
								golog.Info("Remove Expire File: ", f)
							}
						}
					}
				}
			}
		}
	}
	golog.Info("CleanRecordingDir End")
}

func StartTimingTask(name string, isStart bool, regularTime string, taskFunc func()) {
	if isStart {
		golog.Info("Start Timing Task: ", name)
		taskChan := make(chan int)
		go tools.EveryDayTimer(regularTime, taskChan)
		go func() {
			for {
				select {
				case <-taskChan:
					taskFunc()
				default:
					time.Sleep(time.Second * 1)
				}
			}
		}()
	}
}

func GetStreamInfo(streamURL string) (isLive bool, dpi string, bitRate string, fps string) {
	if streamURL == "" {
		return false, "", "", ""
	}
	cmd := exec.Command("ffprobe", "-user_agent", "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36", "-i", streamURL)
	var outb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &outb
	err := cmd.Run()
	if err != nil {
		golog.Error(fmt.Sprintf("GetStreamInfo Error: %s", err.Error()))
		return false, "", "", ""
	}
	reg, _ := regexp.Compile(`Video:.*?(\d+x\d+).*?(\d+) kb\/s.*?(\d+) fps`)
	res := reg.FindAllStringSubmatch(outb.String(), -1)
	if len(res) > 0 {
		return true, res[0][1], res[0][2], res[0][3]
	}
	return false, "", "", ""
}