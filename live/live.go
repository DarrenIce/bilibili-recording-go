package live

import (
	"os/exec"
	"sync"

	"bilibili-recording-go/config"
)

// Live 主类
type Live struct {
	rooms map[string]config.RoomConfigInfo

	stop          chan struct{}
	recordChannel chan config.RoomConfigInfo
	unliveChannel chan string
	decodeChannel chan string
	uploadChannel chan string
	downloadCmd   *exec.Cmd
	state         sync.Map
	lock          *sync.Mutex
}

// Init Live
func (l *Live) Init() {
	l.rooms = make(map[string]config.RoomConfigInfo)
	l.stop = make(chan struct{})
	l.recordChannel = make(chan config.RoomConfigInfo)
	l.unliveChannel = make(chan string)
	l.decodeChannel = make(chan string)
	l.uploadChannel = make(chan string)
	l.lock = new(sync.Mutex)

	go l.recordWorker()
	go l.decodeWorker()

}

// AddRoom ADD
func (l *Live) AddRoom(info config.RoomConfigInfo) {
	l.recordChannel <- info
}
