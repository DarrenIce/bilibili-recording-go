package live

import (
	"fmt"
	"os/exec"
	"time"

	"bilibili-recording-go/infos"
	"bilibili-recording-go/tools"

	"github.com/kataras/golog"
)

func (l *Live) uploadWorker() {
	for {
		roomID := <-l.uploadChannel
		if l.CompareAndSwapUint32(roomID, uploadWait, uploading) {
			infs := infos.New()
			infs.RoomInfos[roomID].UploadStartTime = time.Now().Unix()
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 开始上传", infs.RoomInfos[roomID].Uname, roomID))
			l.Upload(roomID)
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 结束上传", infs.RoomInfos[roomID].Uname, roomID))
			infs.RoomInfos[roomID].UploadEndTime = time.Now().Unix()
			l.CompareAndSwapUint32(roomID, uploading, uploadEnd)
			l.CompareAndSwapUint32(roomID, uploadEnd, start)
		}
	}
}

// Upload upload
func (l *Live) Upload(roomID string) {
	infs := infos.New()
	uname := infs.RoomInfos[roomID].Uname
	uploadName := infs.RoomInfos[roomID].UploadName
	filePath := infs.RoomInfos[roomID].FilePath
	cmd := exec.Command("python", "./live/upload.py", uname, roomID, uploadName, filePath)
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	cmd.Start()
	tools.LiveOutput(stdout)
	cmd.Wait()
}
