package bot

import "gitlab.ozon.dev/MShulgin/homework-2/bot/internal/model"

type Session struct {
	state     func(*Session, chan model.OutgoingMessage, PortfolioService)
	msg       string
	vars      map[string]string
	userId    string
	chatId    int64
	messenger string
	accountId int32
}

type SessionStorage interface {
	GetSession(string) (*Session, bool)
	PutSession(string, *Session)
}

type MapSessionStorage struct {
	store map[string]*Session
}

func NewMapSessionStorage() MapSessionStorage {
	return MapSessionStorage{store: make(map[string]*Session)}
}

func (s *MapSessionStorage) GetSession(key string) (*Session, bool) {
	val, ok := s.store[key]
	return val, ok
}

func (s *MapSessionStorage) PutSession(key string, state *Session) {
	s.store[key] = state
}
