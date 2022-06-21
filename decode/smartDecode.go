package decode

import (
	"bilibili-recording-go/live"
	"bilibili-recording-go/tools"
	"fmt"
	"os"
	"strings"

	"github.com/kataras/golog"
)

func SmartDecode(l *live.LiveSnapshot) {
	if !tools.Exists(fmt.Sprintf("./recording/%s", l.Uname)) || !tools.Exists(fmt.Sprintf("./recording/%s/tmp", l.Uname)) {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 文件夹不存在, SmartDecode Exit.", l.Uname, l.UID))
	}
	for _, v := range tools.ListDir(fmt.Sprintf("./recording/%s/tmp", l.Uname)) {
		if ok := strings.HasSuffix(v, ".flv"); !ok {
			continue
		}
		if v == l.TmpFilePath {
			golog.Info(fmt.Sprintf("%s[RoomID: %s] %s为正在录制文件, 已跳过.", l.Uname, l.UID, v))
		}
		_, outputName := GenerateFileName([]string{v}, l)
		if tools.Exists(fmt.Sprintf("./recording/%s/%s.mp4", l.Uname, outputName)) {
			if l.NeedM4a {
				if tools.Exists(fmt.Sprintf("./recording/%s/%s.m4a", l.Uname, outputName)) {
					continue
				} else {
					os.Remove(fmt.Sprintf("./recording/%s/%s.mp4", l.Uname, outputName))
				}
			} else {
				continue
			}
		}
		decodeChan <- &decodeTuple{
			live:        l,
			convertFile: v,
			outputName: outputName,
		}
	}	
}