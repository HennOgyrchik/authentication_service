package main

import (
	"golang.org/x/crypto/bcrypt"
	"sync"
)

type storage struct {
	bcryptCost int
	sync.RWMutex
	accessMap    map[string]tokenInfo //мапа guid - access
	dbCollection *connect
}

type tokenInfo struct {
	token   string
	expTime int64
}

func newStorage(bcryptCost int) *storage {
	return &storage{
		bcryptCost:   bcryptCost,
		accessMap:    make(map[string]tokenInfo),
		dbCollection: dbConn(),
	}
}

func (s *storage) rememberTokens(guid string, access tokenInfo, refresh tokenInfo) error {
	s.Lock()
	defer s.Unlock()
	//надо захэшировать!
	hashToken, err := hashPassword(refresh.token, s.bcryptCost)
	if err != nil {
		return err
	}
	err = s.dbCollection.insertOne(data{
		Guid:    guid,
		Token:   hashToken,
		ExpTime: refresh.expTime,
	})
	if err != nil {
		return err
	}

	s.accessMap[guid] = tokenInfo{token: access.token, expTime: access.expTime}

	return nil
}

// Удаление токена в мапе и БД
func (s *storage) deleteToken(guid string, hash string) error {
	s.Lock()

	delete(s.accessMap, guid)

	err := s.dbCollection.deleteOne(hash)

	s.Unlock()
	return err

}

// найти хэш по guid
func (s *storage) findHash(guid string, token string) (string, error) {
	rows, err := s.dbCollection.find(guid)
	if err != nil {
		return "", err
	}

	for _, row := range rows {
		if сheckPasswordHash(token, row.Token) {
			return row.Token, nil
		}
	}
	return "", ErrNotFound
}

func (s *storage) findOne(hash string) (data, error) {
	return s.dbCollection.findOne(hash)
}

func hashPassword(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

func сheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
