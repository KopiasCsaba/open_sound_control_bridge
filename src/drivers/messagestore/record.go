package messagestore

import (
	"time"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IMessageStoreRecord = &Record{}

type Record struct {
	message   usecaseifs.IOSCMessage
	arrivedAt time.Time
}

func NewMessageStoreRecord(message usecaseifs.IOSCMessage, arrivedAt time.Time) *Record {
	return &Record{message: message, arrivedAt: arrivedAt}
}

func (m *Record) GetMessage() usecaseifs.IOSCMessage {
	return m.message
}

func (m *Record) GetArrivedAt() time.Time {
	return m.arrivedAt
}
