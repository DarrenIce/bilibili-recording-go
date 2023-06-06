package live

import (
	"bilibili-recording-go/config"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
)

func init() {
	registerSite("douyu", &douyu{})
	PRLock.Lock()
	PlatformRooms["douyu"] = make([]string, 0)
	PRLock.Unlock()
	go flushPlatformLives("douyu")
}

type douyu struct {
}

func (s *douyu) Name() string {
	return "斗鱼"
}

func (s *douyu) SetCookies(cookies string) {}

func (s *douyu) GetInfoByRoom(r *Live) SiteInfo {
	req := requests.Requests()
	c := config.New()
	if c.Conf.RcConfig.NeedProxy {
		req.Proxy(c.Conf.RcConfig.Proxy)
	}
	headers := requests.Header{
		"User-Agent":   "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko); Chrome/75.0.3770.100 Mobile Safari/537.36",
		"Content-Type": "application/x-www-form-urlencoded",
	}
	resp, err := req.Get(fmt.Sprintf("https://www.douyu.com/betard/%s", r.RoomID), headers)
	if err != nil {
		return SiteInfo{
			Title: err.Error(),
		}
	}
	data := resp.Text()
	sInfo := SiteInfo{}
	sInfo.LiveStatus = int(gjson.Get(data, "room.show_status").Int())
	if sInfo.LiveStatus == 2 {
		sInfo.LiveStatus = 0
	}
	sInfo.RealID = gjson.Get(data, "room.room_id").String()
	sInfo.LockStatus = 0
	sInfo.Uname = gjson.Get(data, "room.nickname").String()
	sInfo.UID = gjson.Get(data, "room.owner_uid").String()
	sInfo.Title = gjson.Get(data, "room.room_name").String()
	sInfo.AreaName = gjson.Get(data, "game.tag_name").String()

	return sInfo
}

func (s *douyu) GetRoomLiveURL(roomID string) (string, bool) {
	return "", false
}

func (s *douyu) DownloadLive(r *Live) {
	uname := r.Uname
	outputName := r.AreaName + "_" + r.Title + "_" + fmt.Sprint(time.Unix(r.RecordStartTime, 0).Format("20060102150405")) + ".flv"
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 本次录制文件为：%s", r.Uname, r.RoomID, outputName))
	r.TmpFilePath = fmt.Sprintf("./recording/%s/tmp/%s", uname, outputName)
	middle, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp", uname))
	outputFile := fmt.Sprint(middle + "\\" + outputName)
	url := fmt.Sprintf("http://127.0.0.1:1769/douyu/%s", r.RoomID)
	r.downloadCmd = exec.Command("ffmpeg", "-i", url, "-c", "copy", outputFile)
	// r.downloadCmd = exec.Command("streamlink", "-f", "-o", outputFile, s.liveUrl, "best")
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
	time.Sleep(time.Second * 120)
	r.unlive()
}
