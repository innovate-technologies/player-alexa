package main

import (
	"errors"
	"sync"
	"time"
)

var configCache = map[string]config{}
var configCacheMutex = sync.Mutex{}
var configCacheTime = map[string]time.Time{}

func init() {
	go checkTTL()
}

func getConfigForUser(username string) (config, error) {
	configCacheMutex.Lock()
	if entry, exists := configCache[username]; exists {
		configCacheMutex.Unlock()
		return entry, nil
	}
	configCacheMutex.Unlock()

	entry, err := getITFrameConfig(username)
	if err != nil {
		return entry, err
	}

	tunein, err := getTuneIn(username)
	if err != nil {
		return entry, err
	}
	entry.TuneInURL = tunein

	if len(entry.LanguageEntries) == 0 {
		return entry, errors.New("No configuration entry")
	}

	configCacheMutex.Lock()
	configCache[username] = entry
	configCacheTime[username] = time.Now()
	configCacheMutex.Unlock()

	return entry, nil
}

func checkTTL() {
	configCacheMutex.Lock()
	for username, t := range configCacheTime {
		if t.Before(time.Now().Add(-1 * time.Hour)) {
			delete(configCache, username)
			delete(configCacheTime, username)
		}
	}
	configCacheMutex.Unlock()
	time.Sleep(30 * time.Minute)
}
