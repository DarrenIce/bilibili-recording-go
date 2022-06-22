package decode

import (
	"bilibili-recording-go/live"
	"bilibili-recording-go/tools"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kataras/golog"
)

func SmartDecode(l *live.LiveSnapshot) {
	if !tools.Exists(fmt.Sprintf("./recording/%s", l.Uname)) || !tools.Exists(fmt.Sprintf("./recording/%s/tmp", l.Uname)) {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 文件夹不存在, SmartDecode Exit.", l.Uname, l.UID))
	}
	flst, err := ioutil.ReadDir(fmt.Sprintf("./recording/%s/tmp", l.Uname))
	if err != nil {
		golog.Error(err)
		return
	}
	for _, f := range flst {
		if fmt.Sprintf("./recording/%s/tmp/%s", l.Uname, f.Name()) == l.TmpFilePath {
			golog.Info(fmt.Sprintf("%s[RoomID: %s] %s为正在录制文件, 已跳过.", l.Uname, l.UID, f.Name()))
			continue
		}
		if strings.HasSuffix(f.Name(), ".flv") && f.Size() < 1024 * 1024 * 50 {
			golog.Info(fmt.Sprintf("%s[RoomID: %s] %s 文件太小, 已删除.", l.Uname, l.UID, f.Name()))
			os.Remove(fmt.Sprintf("./recording/%s/tmp/%s", l.Uname, f.Name()))
		}
	}
	for _, v := range tools.ListDir(fmt.Sprintf("./recording/%s/tmp", l.Uname)) {
		if ok := strings.HasSuffix(v, ".flv"); !ok {
			continue
		}
		if v == l.TmpFilePath {
			golog.Info(fmt.Sprintf("%s[RoomID: %s] %s为正在录制文件, 已跳过.", l.Uname, l.UID, v))
			continue
		}
		_, outputName := GenerateFileName([]string{v}, l)
		if tools.Exists(fmt.Sprintf("./recording/%s/%s.mp4", l.Uname, outputName)) {
			if l.NeedM4a {
				if tools.Exists(fmt.Sprintf("./recording/%s/%s.m4a", l.Uname, outputName)) {
					golog.Info(fmt.Sprintf("%s[RoomID: %s] %s 已转码完成, 删除tmp文件.", l.Uname, l.UID, v))
					os.Remove(v)
					continue
				} else {
					golog.Info(fmt.Sprintf("%s[RoomID: %s] %s转码未完成，重新加入队列.", l.Uname, l.UID, outputName))
					os.Remove(fmt.Sprintf("./recording/%s/%s.mp4", l.Uname, outputName))
				}
			} else {
				golog.Info(fmt.Sprintf("%s[RoomID: %s] %s 已转码完成, 删除tmp文件.", l.Uname, l.UID, v))
				os.Remove(v)
				continue
			}
		}
		golog.Info(fmt.Sprintf("%s[RoomID: %s] %s 加入转码队列.", l.Uname, l.UID, v))
		decodeChan <- &decodeTuple{
			live:        l,
			convertFile: v,
			outputName: outputName,
		}
	}	
}