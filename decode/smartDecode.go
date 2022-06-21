package decode

import (
	"bilibili-recording-go/live"
	"bilibili-recording-go/tools"
	"fmt"

	"github.com/kataras/golog"
)

func SmartDecode(l *live.LiveSnapshot) {
	if !tools.Exists(fmt.Sprintf("./recording/%s", l.Uname)) || !tools.Exists(fmt.Sprintf("./recording/%s/tmp", l.Uname)) {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 文件夹不存在, SmartDecode Exit.", l.Uname, l.UID))
	}
	for f := range tools.ListDir(fmt.Sprintf("./recording/%s/tmp", l.Uname)) {
		fmt.Println(f)
	}	
}