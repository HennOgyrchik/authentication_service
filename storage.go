package main

import (
	"strconv"
	"sync"
	"time"
)

type storage struct {
	sync.RWMutex
	accessMap    map[string][2]string //мапа guid - access token
	dbCollection *connect
}

type userInfo struct {
	guid                string
	accessToken         string
	expTimeAccessToken  time.Duration
	refreshToken        string
	expTimeRefreshToken time.Duration
}

func newStorage() *storage {
	return &storage{
		accessMap:    make(map[string][2]string),
		dbCollection: dbConn(),
	}
}

func (s *storage) rememberTokens(user userInfo) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.dbCollection.insertOne(user.refreshToken, user.expTimeRefreshToken)
	if err != nil {
		return err
	}

	s.accessMap[user.guid] = [2]string{user.accessToken, strconv.FormatInt(time.Now().Unix()+int64(user.expTimeAccessToken), 10)}

	return nil
}
