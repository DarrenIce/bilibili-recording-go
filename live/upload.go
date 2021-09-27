package live

import (
	"fmt"
	"os/exec"
	"time"
	"sync/atomic"

	"bilibili-recording-go/tools"

	"github.com/kataras/golog"
)

func uploadWorker() {
	for {
		roomID := <- uploadChan
		if _, ok := Lives[roomID]; !ok {
			continue
		}
		LmapLock.Lock()
		live := Lives[roomID]
		LmapLock.Unlock()
		if atomic.CompareAndSwapUint32(&live.State, uploadWait, uploading) {
			live.UploadStartTime = time.Now().Unix()
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 开始上传", live.Uname, roomID))
			live.Upload()
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 结束上传", live.Uname, roomID))
			live.UploadEndTime = time.Now().Unix()
			atomic.CompareAndSwapUint32(&live.State, uploading, uploadEnd)
			atomic.CompareAndSwapUint32(&live.State, uploadEnd, start)
		}
	}
}

// Upload upload
func (l *Live) Upload() {
	uname := l.Uname
	uploadName := l.UploadName
	filePath := l.FilePath
	cmd := exec.Command("python", "./live/upload.py", uname, l.RoomID, uploadName, filePath)
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	cmd.Start()
	tools.LiveOutput(stdout)
	cmd.Wait()
}
