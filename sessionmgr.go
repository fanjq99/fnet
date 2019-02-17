package fnet

import (
	"sync"
	"sync/atomic"

	"github.com/fanjq99/glog"
)

type SessionManager struct {
	sync.RWMutex
	sessionMap map[int64]ISession
	accID      int64
}

const totalTryCount int = 100

func (sm *SessionManager) Add(session ISession) {
	sm.Lock()
	defer sm.Unlock()

	var tryCount = totalTryCount
	var id int64

	for tryCount > 0 {
		id = atomic.AddInt64(&sm.accID, 1)
		glog.Info("Session add id", id)
		if _, ok := sm.sessionMap[id]; !ok {
			// 找到一个新的id
			break
		}
		tryCount--
	}

	if tryCount == 0 {
		glog.Errorln("sessionID override!", id)
	}

	session.SetID(id)
	sm.sessionMap[id] = session
	glog.Infof("Add Total connection: %d", len(sm.sessionMap))
}

func (sm *SessionManager) Remove(session ISession) {
	sm.Lock()
	defer sm.Unlock()

	_, ok := sm.sessionMap[session.GetID()]
	if ok {
		delete(sm.sessionMap, session.GetID())
		glog.Infof("Remove Total connection: %d", len(sm.sessionMap))
	}
}

func (sm *SessionManager) Get(sid int64) ISession {
	sm.RLock()
	defer sm.RUnlock()

	if v, ok := sm.sessionMap[sid]; ok {
		return v
	}

	return nil
}

func (sm *SessionManager) CloseAll() {
	tmp := make([]ISession, 0, sm.Len())

	sm.Lock()
	for _, v := range sm.sessionMap {
		tmp = append(tmp, v)
	}
	sm.sessionMap = make(map[int64]ISession)
	sm.Unlock()

	for _, v := range tmp {
		v.Close()
	}
}

func (sm *SessionManager) Len() int {
	sm.RLock()
	defer sm.RUnlock()
	return len(sm.sessionMap)
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessionMap: make(map[int64]ISession),
		accID:      0,
	}
}
