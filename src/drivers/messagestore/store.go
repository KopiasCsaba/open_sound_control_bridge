// Package messagestore implements the core of the application, the store that accepts OSC messages,
// and on which the conditions run.
package messagestore

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"net.kopias.oscbridge/app/pkg/slicetools"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IMessageStore = &MessageStore{}

type MessageStore struct {
	m                     *sync.RWMutex
	store                 map[string]usecaseifs.IMessageStoreRecord
	watchedRecord         *usecaseifs.IOSCMessage
	watchedRecordAccesses int64
}

// WatchRecordAccess registers a message, and from the point of the call, the store will cound how many times that address has been accessed.
// See [MessageStore.GetWatchedRecordAccesses].
func (e *MessageStore) WatchRecordAccess(msg *usecaseifs.IOSCMessage) {
	e.watchedRecord = msg
	e.watchedRecordAccesses = 0
}

// GetWatchedRecordAccesses returns the number of accesses that the watched record received.
// See [MessageStore.GetWatchedRecordAccesses]
func (e *MessageStore) GetWatchedRecordAccesses() int64 {
	return e.watchedRecordAccesses
}

// checkWatchedRecordAccess determines if the list of about-to-return messages [msg] contains the watched message.
// [trackAccess] can override watching altogether.
func (e *MessageStore) checkWatchedRecordAccess(msgs []usecaseifs.IOSCMessage, trackAccess bool) {
	if !trackAccess {
		return
	}

	if e.watchedRecord == nil {
		return
	}

	for _, m := range msgs {
		if m.Equal(*e.watchedRecord) {
			e.watchedRecordAccesses++
		}
	}
}

// Clone clones this message store
func (e *MessageStore) Clone() usecaseifs.IMessageStore {
	newStore := NewMessageStore()
	newStore.store = e.GetAll()

	return newStore
}

// GetAll returns every record. It does not do record watching.
func (e *MessageStore) GetAll() map[string]usecaseifs.IMessageStoreRecord {
	newData := map[string]usecaseifs.IMessageStoreRecord{}

	e.m.RLock()
	for key, value := range e.store {
		newData[key] = value
	}
	// we clone ourselves, hopefully will always work
	// newData := deepcopy.Anything(e.store).(map[string]usecaseifs.IMessageStoreRecord)
	e.m.RUnlock()

	return newData
}

// GetRecord returns a message with the exact [address].
// [trackAccess] determines if this access is tracked or not. See WatchRecordAccess.
func (e *MessageStore) GetRecord(address string, trackAccess bool) (usecaseifs.IMessageStoreRecord, bool) {
	e.m.RLock()
	record, ok := e.store[address]
	if ok {
		e.checkWatchedRecordAccess([]usecaseifs.IOSCMessage{record.GetMessage()}, trackAccess)
	}
	e.m.RUnlock()

	return record, ok
}

// GetOneRecordByRegexp returns a single record whose address is matching the [re] expression.
// [trackAccess] determines if this access is tracked or not. See WatchRecordAccess.
func (e *MessageStore) GetOneRecordByRegexp(re string, trackAccess bool) (usecaseifs.IMessageStoreRecord, error) {
	result, err := e.getRecordsByRegexpFinder(re, true, trackAccess)

	if len(result) == 0 {
		return nil, err
	}

	e.checkWatchedRecordAccess([]usecaseifs.IOSCMessage{result[0].GetMessage()}, trackAccess)
	return result[0], err
}

// GetRecordsByRegexp returns a list of records whose address is matching the [re] expression.
// [trackAccess] determines if this access is tracked or not. See WatchRecordAccess.
func (e *MessageStore) GetRecordsByRegexp(re string, trackAccess bool) ([]usecaseifs.IMessageStoreRecord, error) {
	return e.getRecordsByRegexpFinder(re, false, trackAccess)
}

// getRecordsByRegexpFinder does the grunt work for GetRecordsByRegexp and GetOneRecordByRegexp.
func (e *MessageStore) getRecordsByRegexpFinder(re string, firstOnly bool, trackAccess bool) ([]usecaseifs.IMessageStoreRecord, error) {
	compiledRegex, err := regexp.Compile(re)
	if err != nil {
		return nil, err
	}

	result := []usecaseifs.IMessageStoreRecord{}

	e.m.RLock()
	for address, record := range e.store {
		if compiledRegex.MatchString(address) {
			result = append(result, record)
			if firstOnly {
				break
			}
		}
	}
	e.checkWatchedRecordAccess(slicetools.Map(result, func(t usecaseifs.IMessageStoreRecord) usecaseifs.IOSCMessage {
		return t.GetMessage()
	}), trackAccess)

	e.m.RUnlock()
	return result, nil
}

// GetRecordsByPrefix returns a list of records whose address starts with [prefix].
// [trackAccess] determines if this access is tracked or not. See WatchRecordAccess.
func (e *MessageStore) GetRecordsByPrefix(prefix string, trackAccess bool) []usecaseifs.IMessageStoreRecord {
	result := []usecaseifs.IMessageStoreRecord{}

	e.m.RLock()
	for address, record := range e.store {
		if strings.HasPrefix(address, prefix) {
			result = append(result, record)
		}
	}
	e.checkWatchedRecordAccess(slicetools.Map(result, func(t usecaseifs.IMessageStoreRecord) usecaseifs.IOSCMessage {
		return t.GetMessage()
	}), trackAccess)
	e.m.RUnlock()
	return result
}

// SetRecord updates the store. Returns the fact if the store changed or not.
func (e *MessageStore) SetRecord(record usecaseifs.IOSCMessage) bool {
	changed := false
	e.m.Lock()

	oldRecord, ok := e.store[record.GetAddress()]

	if ok && !oldRecord.GetMessage().Equal(record) || !ok {
		e.store[record.GetAddress()] = NewMessageStoreRecord(record, time.Now())
		changed = true
	}

	e.m.Unlock()
	return changed
}

func NewMessageStore() *MessageStore {
	return &MessageStore{
		m:     &sync.RWMutex{},
		store: make(map[string]usecaseifs.IMessageStoreRecord),
	}
}
