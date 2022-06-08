package live

import "sync"

var sites sync.Map

type SiteInfo struct {
	RealID        string
	LiveStatus    int
	LockStatus    int
	Uname         string
	UID           string
	Title         string
	LiveStartTime int64
	AreaName      string
}

type Site interface {
	Name() string
	GetInfoByRoom(*Live) SiteInfo
	DownloadLive(*Live)
}

func registerSite(siteID string, site Site) {
	if _, dup := sites.LoadOrStore(siteID, site); dup {
		panic("site already registered.")
	}
}

func Sniff(siteID string) (Site, bool) {
	s, ok := sites.Load(siteID)
	if !ok {
		return nil, ok
	}
	return s.(Site), ok
}