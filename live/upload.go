package live

import (
	"fmt"
	"os/exec"
	"time"

	"bilibili-recording-go/infos"
	"bilibili-recording-go/tools"
	"bilibili-recording-go/config"

	"github.com/kataras/golog"
)

func (l *Live) uploadWorker() {
	for {
		roomID := <-l.uploadChannel
		if l.compareAndSwapUint32(roomID, updateWait, updating) {
			infs := infos.New()
			infs.RoomInfos[roomID].UploadStartTime = time.Now().Unix()
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 开始上传", infs.RoomInfos[roomID].Uname, roomID))
			l.Upload(roomID)
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 结束上传", infs.RoomInfos[roomID].Uname, roomID))
			infs.RoomInfos[roomID].UploadEndTime = time.Now().Unix()
			l.compareAndSwapUint32(roomID, updating, updateEnd)
			l.compareAndSwapUint32(roomID, updateEnd, start)
		}
	}
}

// Upload upload
func (l *Live) Upload(roomID string) {
	infs := infos.New()
	uname := infs.RoomInfos[roomID].Uname
	uploadName := infs.RoomInfos[roomID].UploadName
	filePath := infs.RoomInfos[roomID].FilePath
	c := config.New()
	cookies, _ := tools.LoginByPassword(c.Conf.Bili.User, c.Conf.Bili.Password)
	DedeUserID := cookies["DedeUserID"]
	DedeUserID__ckMd5 := cookies["DedeUserID__ckMd5"]
	SESSDATA := cookies["SESSDATA"]
	bili_jct := cookies["bili_jct"]
	sid := cookies["sid"]
	cmd := exec.Command("python", "./live/upload.py", uname, roomID, uploadName, filePath, DedeUserID, DedeUserID__ckMd5, SESSDATA, bili_jct, sid)
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	cmd.Start()
	tools.LiveOutput(stdout)
	cmd.Wait()
}
