package logging

import (
	"log"
	"sync"
)

type MemorySessionStore struct {
	Sessions map[string]*LogSession
	Lock     sync.RWMutex
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		Sessions: make(map[string]*LogSession),
	}
}

func (s *MemorySessionStore) Get(id string) *LogSession {
	s.Lock.RLock()
	defer s.Lock.RUnlock()
	return s.Sessions[id]
}

func (s *MemorySessionStore) Set(id string, session *LogSession) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.Sessions[id] = session
}

func (s *MemorySessionStore) Close(id, reason string, status uint32) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	sess, _ := s.Sessions[id]
	err := sess.sockJSSession.Close(status, reason)
	if err != nil {
		log.Println(err)
	}
	delete(s.Sessions, id)
}

func (s *MemorySessionStore) Clean() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	for _, v := range s.Sessions {
		v.sockJSSession.Close(2, "system is logout, please retry...")
	}
	s.Sessions = make(map[string]*LogSession)
}
