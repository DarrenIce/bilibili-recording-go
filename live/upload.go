package live

import (
	"fmt"
	"time"

	"bilibili-recording-go/infos"

	"github.com/kataras/golog"
)


func (l *Live) uploadWorker() {
	for {
		roomID := <- l.uploadChannel
		if l.compareAndSwapUint32(roomID, updateWait, updating) {
			infs := infos.New()
			infs.RoomInfos[roomID].UploadStartTime = time.Now().Format("2006-01-02 15:04:05")
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 开始上传", infs.RoomInfos[roomID].Uname, roomID))
			l.Upload(roomID)
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 结束上传", infs.RoomInfos[roomID].Uname, roomID))
			infs.RoomInfos[roomID].UploadEndTime = time.Now().Format("2006-01-02 15:04:05")
			l.compareAndSwapUint32(roomID, updating, updateEnd)
			l.compareAndSwapUint32(roomID, updateEnd, start)
		}
	}
}

// Upload upload
func (l *Live) Upload(roomID string) {

}