package common

import "sync"

// SyncMap is a map synchronized using RWMutex
type SyncMap struct {
	baseMap map[string]interface{}

	mutex sync.RWMutex
}

// Creates new sync map
func NewSyncMap() *SyncMap {
	return &SyncMap{
		baseMap: make(map[string]interface{}),
	}
}

// Store new value
func (sm *SyncMap) Store(id string, value interface{}) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.baseMap[id] = value
}

// Load value by id which is the key value
func (sm *SyncMap) Load(id string) (interface{}, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	value, ok := sm.baseMap[id]
	return value, ok
}

// Delete value by key
func (sm *SyncMap) Delete(id string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	delete(sm.baseMap, id)
}

// Lenght of the map
func (sm *SyncMap) Lenght() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return len(sm.baseMap)
}
