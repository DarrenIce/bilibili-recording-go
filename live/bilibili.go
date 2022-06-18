package live

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"

	"bilibili-recording-go/config"
	"bilibili-recording-go/tools"
)

func init() {
	registerSite("bilibili", &bilibili{})
}

type bilibili struct {
	liveUrl string
}

func (s *bilibili) Name() string {
	return "哔哩哔哩"
}

func (s *bilibili) SetCookies(cookies string) {
}

func (s *bilibili) GetInfoByRoom(r *Live) SiteInfo {
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
		return r.SiteInfo
	}
	data := gjson.Get(resp.Text(), "data")
	liveStatus := int(data.Get("room_info").Get("live_status").Int())
	if liveStatus == 1 {
		url := fmt.Sprintf("http://api.live.bilibili.com/room/v1/Room/playUrl?cid=%s&quality=4", r.RoomID)
		req := requests.Requests()
		headers := requests.Header{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66",
		}
		resp, _ := req.Get(url, headers)
		data := gjson.Get(resp.Text(), "data")
		if resp.R.StatusCode == 200 {
			s.liveUrl = data.Get("durl.0.url").String()
		} else {
			fmt.Printf("%s[RoomID: %s] 获取直播地址失败，请检查直播状态\n", r.Uname, r.RoomID)
		}
	}
	if resp.R.StatusCode == 200 {
		return SiteInfo{
			RealID:        data.Get("room_info").Get("room_id").String(),
			LiveStatus:    liveStatus,
			LockStatus:    int(data.Get("room_info").Get("lock_status").Int()),
			Uname:         data.Get("anchor_info").Get("base_info").Get("uname").String(),
			UID:           data.Get("room_info").Get("uid").String(),
			Title:         data.Get("room_info").Get("title").String(),
			LiveStartTime: data.Get("room_info").Get("live_start_time").Int(),
			AreaName:      data.Get("room_info").Get("area_name").String(),
		}
	} else {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " 412啦, 快换代理")
		return r.SiteInfo
	}
}

func (s *bilibili) DownloadLive(r *Live) {
	isLive, dpi, bitRate, fps := GetStreamInfo(s.liveUrl)
	if !isLive {
		fmt.Printf("%s[RoomID: %s] 直播状态不正常\n", r.Uname, r.RoomID)
		atomic.CompareAndSwapUint32(&r.State, running, start)
		return
	}
	uname := r.Uname
	tools.Mkdir(fmt.Sprintf("./recording/%s/tmp", uname))
	outputName := r.AreaName + "_" + r.Title + "_" + fmt.Sprint(time.Unix(r.RecordStartTime, 0).Format("20060102150405")) + ".flv"
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 本次录制文件为：%s, 分辨率: %s, 码率: %s, fps: %s", r.Uname, r.RoomID, outputName, dpi, bitRate, fps))
	r.TmpFilePath = fmt.Sprintf("./recording/%s/tmp/%s", uname, outputName)
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
