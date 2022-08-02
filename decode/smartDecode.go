package decode

import (
	"bilibili-recording-go/live"
	"bilibili-recording-go/tools"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	_ "time"

	"github.com/kataras/golog"
)

func SmartDecode(l *live.LiveSnapshot) {
	if !tools.Exists(fmt.Sprintf("./recording/%s", l.Uname)) || !tools.Exists(fmt.Sprintf("./recording/%s/tmp", l.Uname)) {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 文件夹不存在, SmartDecode Exit.", l.Uname, l.RoomID))
	}
	flst, err := ioutil.ReadDir(fmt.Sprintf("./recording/%s/tmp", l.Uname))
	for k, f := range flst {
		if !strings.HasSuffix(f.Name(), ".flv") {
			if k + 1 < len(flst) {
				flst = append(flst[:k], flst[k+1:]...)
			} else {
				flst = flst[:k]
			}
		}
	}
	sort.Slice(flst, func(i, j int) bool {
		iTime, _ := strconv.ParseInt(strings.Split(strings.TrimSuffix(flst[i].Name(), ".flv"), "_")[2], 10, 64)
		jTime, _ := strconv.ParseInt(strings.Split(strings.TrimSuffix(flst[j].Name(), ".flv"), "_")[2], 10, 64)
		return iTime < jTime
	})
	if err != nil {
		golog.Error(err)
		return
	}
	for _, f := range flst {
		if fmt.Sprintf("./recording/%s/tmp/%s", l.Uname, f.Name()) == l.TmpFilePath {
			golog.Info(fmt.Sprintf("%s[RoomID: %s] %s为正在录制文件, 已跳过.", l.Uname, l.RoomID, f.Name()))
			continue
		}
		// if strings.HasSuffix(f.Name(), ".flv") && f.Size() < 1024 * 1024 * 50 {
		// 	golog.Info(fmt.Sprintf("%s[RoomID: %s] %s 文件太小, 已删除.", l.Uname, l.RoomID, f.Name()))
		// 	os.Remove(fmt.Sprintf("./recording/%s/tmp/%s", l.Uname, f.Name()))
		// }
	}
	for _, v := range tools.ListDir(fmt.Sprintf("./recording/%s/tmp", l.Uname)) {
		if ok := strings.HasSuffix(v, ".flv"); !ok {
			continue
		}
		if v == l.TmpFilePath {
			golog.Info(fmt.Sprintf("%s[RoomID: %s] %s为正在录制文件, 已跳过.", l.Uname, l.RoomID, v))
			continue
		}
		_, outputName := GenerateFileName([]string{v}, l)
		if tools.Exists(fmt.Sprintf("./recording/%s/%s.mp4", l.Uname, outputName)) {
			if l.NeedM4a {
				if tools.Exists(fmt.Sprintf("./recording/%s/%s.m4a", l.Uname, outputName)) {
					golog.Info(fmt.Sprintf("%s[RoomID: %s] %s 已转码完成, 删除tmp文件.", l.Uname, l.RoomID, v))
					os.Remove(v)
					continue
				} else {
					golog.Info(fmt.Sprintf("%s[RoomID: %s] %s转码未完成，重新加入队列.", l.Uname, l.RoomID, outputName))
					os.Remove(fmt.Sprintf("./recording/%s/%s.mp4", l.Uname, outputName))
				}
			} else {
				golog.Info(fmt.Sprintf("%s[RoomID: %s] %s 已转码完成, 删除tmp文件.", l.Uname, l.RoomID, v))
				os.Remove(v)
				continue
			}
		}
		golog.Info(fmt.Sprintf("%s[RoomID: %s] %s 加入转码队列.", l.Uname, l.RoomID, v))
		decodeChan <- &decodeTuple{
			live:        l,
			convertFile: v,
			outputName: outputName,
		}
	}	
}