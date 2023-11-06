package osc_message

import (
	"fmt"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IOSCMessageArgument = &MessageArgument{}

type MessageArgument struct {
	msgType  string
	msgValue string
}

func NewMessageArgument(msgType string, msgValue string) *MessageArgument {
	return &MessageArgument{msgType: msgType, msgValue: msgValue}
}

func (m *MessageArgument) GetType() string {
	return m.msgType
}

func (m *MessageArgument) GetValue() string {
	return m.msgValue
}

func (m *MessageArgument) String() string {
	return fmt.Sprintf("Argument(%s:%s)", m.msgType, m.msgValue)
}
