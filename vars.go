package fastapi

import (
	"sync"
	"time"
)

var (
	globalMode Runmode
	accessMap  *AccessMap
)

func init() {
	accessMap = newAccessMap()

	setLogger()
	SetLang(Chinese)
}

func SetMode(mode Runmode) {
	globalMode = mode
}

func GetMode() Runmode {
	return globalMode
}

var ContentType = struct {
	Text string
	JSON string
	Form string
}{
	Text: "text/plain",
	JSON: "application/json",
	Form: "application/x-www-form-urlencoded",
}

func newAccessMap() *AccessMap {
	o := &AccessMap{
		RWMutex: &sync.RWMutex{},
		data:    make(map[string]int64),
	}

	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			<-ticker.C
			o.Lock()
			o.data = make(map[string]int64)
			o.Unlock()
		}
	}()

	return o
}

type AccessMap struct {
	*sync.RWMutex
	data map[string]int64
}

func (a *AccessMap) Add(p string) {
	a.Lock()
	n, ok := a.data[p]
	if !ok {
		a.data[p] = 1
	} else {
		a.data[p] = n + 1
	}
	a.Unlock()
}

func (a *AccessMap) Sub(p string) {
	a.Lock()
	n, ok := a.data[p]
	if ok {
		a.data[p] = n - 1
	}
	a.Unlock()
}

func (a *AccessMap) Get(p string) int64 {
	a.RLock()
	defer a.RUnlock()
	return a.data[p]
}
