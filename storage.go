package main

import (
	"github.com/Microsoft/go-winio/pkg/guid"
	"sync"
)

type storage struct {
	sync.RWMutex
	accessMap map[guid.GUID]string //мапа guid - access token
}

// ДОДЕЛАТЬ! не забыть удалить из мапы если при записи в БД произошла ошибка
func (s *storage) rememberTokens(guid guid.GUID, accessToken string, refreshToken string) error {
	s.Lock()
	s.accessMap[guid] = accessToken
	s.Unlock()

	//тутнадо положить refresh в монгу!

	return nil
}
