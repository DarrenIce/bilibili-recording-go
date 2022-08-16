package live

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"

	"bilibili-recording-go/config"
)

func init() {
	registerSite("bigo", &bigo{})
	PRLock.Lock()
	PlatformRooms["bigo"] = make([]string, 0)
	PRLock.Unlock()
	go flushPlatformLives("bigo")
}

type bigo struct {
	liveUrl string
}

func (s *bigo) Name() string {
	return "BIGO"
}

func (s *bigo) SetCookies(cookies string) {
}

func (s *bigo) GetInfoByRoom(r *Live) SiteInfo {
	defer func() {
		if v := recover(); v != nil {
			golog.Error("捕获了一个恐慌: ", v)
			return
		}
	}()
	url := fmt.Sprintf("https://ta.bigo.tv/official_website/studio/getInternalStudioInfo?siteId=%s", r.RoomID)
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
	resp, err := req.Post(url, headers)
	if err != nil {
		golog.Error(err)
		return r.SiteInfo
	}
	data := gjson.Get(resp.Text(), "data")
	liveStatus := int(data.Get("alive").Int())
	if liveStatus == 1 {
		s.liveUrl = data.Get("hls_src").String()
	}
	if resp.R.StatusCode == 200 {
		return SiteInfo{
			RealID:        data.Get("siteId").String(),
			LiveStatus:    liveStatus,
			LockStatus:    0,
			Uname:         data.Get("nick_name").String(),
			UID:           data.Get("uid").String(),
			Title:         data.Get("roomTopic").String(),
			LiveStartTime: time.Now().Unix(),
			AreaName:      data.Get("country_code").String(),
		}
	} else {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " 412啦, 快换代理")
		return r.SiteInfo
	}
}

func (s *bigo) GetRoomLiveURL(roomID string) (string, bool) {
	url := fmt.Sprintf("https://ta.bigo.tv/official_website/studio/getInternalStudioInfo?siteId=%s", roomID)
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
	resp, err := req.Post(url, headers)
	if err != nil {
		golog.Error(err)
		return "", false
	}
	data := gjson.Get(resp.Text(), "data")
	liveStatus := int(data.Get("alive").Int())
	if liveStatus == 1 {
		return data.Get("hls_src").String(), true
	}
	return "", false
}

func (s *bigo) DownloadLive(r *Live) {
	uname := r.Uname
	outputName := r.AreaName + "_" + r.Title + "_" + fmt.Sprint(time.Unix(r.RecordStartTime, 0).Format("20060102150405")) + ".flv"
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 本次录制文件为：%s", r.Uname, r.RoomID, outputName))
	r.TmpFilePath = fmt.Sprintf("./recording/%s/tmp/%s", uname, outputName)
	middle, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp", uname))
	outputFile := fmt.Sprint(middle + "\\" + outputName)
	r.downloadCmd = exec.Command("ffmpeg", "-i", s.liveUrl, "-c", "copy", outputFile)
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
