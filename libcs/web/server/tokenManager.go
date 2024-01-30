package server

import (
	"errors"
	"github.com/isrc-cas/gt/util"
	"sync"
	"time"
)

type TempKeyDetails struct {
	ExpireAt    time.Time
	UseCount    int
	MaxUse      int
	ActualToken string
}
type TokenManagerConfig struct {
	MaxUse         int
	TempKeyLen     int
	ExpireDuration time.Duration
}

type TokenManager struct {
	tempKeys map[string]*TempKeyDetails
	config   TokenManagerConfig
	mtx      sync.Mutex
}

func NewTokenManager(config TokenManagerConfig) *TokenManager {
	return &TokenManager{
		tempKeys: make(map[string]*TempKeyDetails),
		config:   config,
	}
}
func GetTokenManagerConfig(maxUse int, tempKeyLen int, expireDuration time.Duration) TokenManagerConfig {
	return TokenManagerConfig{
		MaxUse:         maxUse,
		TempKeyLen:     tempKeyLen,
		ExpireDuration: expireDuration,
	}
}

func DefaultTokenManagerConfig() TokenManagerConfig {
	return GetTokenManagerConfig(3, 32, 3*time.Minute)
}

func (tm *TokenManager) GenerateTempKey(actualToken string) string {
	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	tempKey := tm.generateTempKey()
	tm.tempKeys[tempKey] = &TempKeyDetails{
		ExpireAt:    time.Now().Add(tm.config.ExpireDuration),
		UseCount:    0,
		MaxUse:      tm.config.MaxUse,
		ActualToken: actualToken,
	}
	return tempKey
}

func (tm *TokenManager) GetActualToken(tempKey string) (string, error) {
	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	detail, exists := tm.tempKeys[tempKey]
	if !exists {
		return "", errors.New("temp key does not exist")
	}

	if detail.ExpireAt.Before(time.Now()) {
		delete(tm.tempKeys, tempKey) // remove expired tempKey
		return "", errors.New("temp key has expired")
	}

	if detail.UseCount >= detail.MaxUse {
		delete(tm.tempKeys, tempKey) // remove tempKey that has reached max usage
		return "", errors.New("temp key has exceeded its maximum usage")
	}

	detail.UseCount++
	return detail.ActualToken, nil
}

func (tm *TokenManager) generateTempKey() string {
	return util.RandomString(tm.config.TempKeyLen)
}
