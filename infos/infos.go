package infos

import (
	"net/http"
	"sync"
)

type biliInfo struct {
	Username string
	Password string
	Cookies  []*http.Cookie
}

// LiveInfos liveinfos
type LiveInfos struct {
	BiliInfo  *biliInfo

	lock *sync.Mutex
}

var (
	once sync.Once

	instance *LiveInfos
)

// New new
func New() *LiveInfos {
	once.Do(func() {
		instance = new(LiveInfos)
		instance.init()
	})

	return instance
}

func (l *LiveInfos) init() {
	l.BiliInfo = new(biliInfo)
	l.lock = new(sync.Mutex)
}

