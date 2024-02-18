package storage

import (
	"golang.org/x/crypto/bcrypt"
	"medods/internal/storage/mongo"
	"sync"
)

type Storage struct {
	bcryptCost int
	sync.RWMutex
	accessMap    map[string]*TokenInfo //мапа guid - access
	dbCollection *mongo.Connect
}

type TokenInfo struct {
	Token   string
	ExpTime int64
	GUID    string
}

// NewStorage создание нового хранилища для Access токенов и подключения к БД для Refresh токена
func NewStorage(bcryptCost int, dbAddr string, dbPort string) (*Storage, error) {
	conn, err := mongo.DBConn(dbAddr, dbPort)
	if err != nil {
		return nil, err
	}
	return &Storage{
		bcryptCost:   bcryptCost,
		accessMap:    make(map[string]*TokenInfo),
		dbCollection: conn,
	}, err
}

// RememberTokens запись Access токена в мапу, а Refresh в БД
func (s *Storage) RememberTokens(access TokenInfo, refresh TokenInfo) error {
	s.Lock()
	defer s.Unlock()

	hashToken, err := hashPassword(refresh.Token, s.bcryptCost)
	if err != nil {
		return err
	}
	err = s.dbCollection.InsertOne(refresh.GUID, hashToken, refresh.ExpTime)
	if err != nil {
		return err
	}

	s.accessMap[refresh.GUID] = &TokenInfo{Token: access.Token, ExpTime: access.ExpTime}

	return nil
}

// DeleteToken удаление токенов в мапе и БД
func (s *Storage) DeleteToken(guid string, hash string) error {
	s.Lock()

	delete(s.accessMap, guid)

	err := s.dbCollection.DeleteOne(hash)

	s.Unlock()
	return err
}

// FindHash поиск хэша токена по guid в БД
func (s *Storage) FindHash(guid string, token string) (string, error) {
	rows, err := s.dbCollection.Find(guid)
	if err != nil {
		return "", err
	}

	for _, row := range rows {
		if checkPasswordHash(token, row.Token) {
			return row.Token, nil
		}
	}
	return "", ErrNotFound
}

// hashPassword хэширование строки
func hashPassword(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// checkPasswordHash проверка хэша на соответствие предполагаемой строке
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ClearStorage очистка хранилищ от данных
func (s *Storage) ClearStorage() error {
	s.accessMap = make(map[string]*TokenInfo)
	return s.dbCollection.Drop()
}

func (s *Storage) FindOneInDB(token string) (TokenInfo, error) {
	data, err := s.dbCollection.FindOne(token)

	return TokenInfo{
		Token:   data.Token,
		ExpTime: data.ExpTime,
		GUID:    data.Guid,
	}, err
}
