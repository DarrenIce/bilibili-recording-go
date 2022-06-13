package live

import (
	"bilibili-recording-go/config"
	"fmt"
	"regexp"

	"github.com/asmcos/requests"
)

func init() {
	registerSite("huya", &huya{})
}

type huya struct {}

func (s *huya) Name() string {
	return "虎牙"
}

func (s *huya) SetCookies(cookies string) {}

func (s * huya) GetInfoByRoom(r *Live) SiteInfo {
	req := requests.Requests()
	c := config.New()
	if c.Conf.RcConfig.NeedProxy {
		req.Proxy(c.Conf.RcConfig.Proxy)
	}
	headers := requests.Header{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.71 Safari/537.36 Edg/94.0.992.38",
		"Content-Type": "application/x-www-form-urlencoded",
	}
	resp, err := req.Get(fmt.Sprintf("https://m.huya.com/%s", r.RoomID), headers)
	if err != nil {
		return SiteInfo{
			Title: err.Error(),
		}
	}
	re := regexp.MustCompile(`window.HNF_GLOBAL_INIT = ({.*})\s+<\/script>`)
	data := re.FindStringSubmatch(resp.Text())
	fmt.Println(data)
	return SiteInfo{
		Title: "暂不支持",
	}
}

func (s *huya)DownloadLive(r *Live) {
}