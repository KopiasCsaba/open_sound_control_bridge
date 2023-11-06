package osc_message

import (
	"fmt"
	"strings"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IOSCMessage = &Message{}

// Message is the internal implementation of IOSCMessage
type Message struct {
	address   string
	arguments []usecaseifs.IOSCMessageArgument
}

func NewMessage(address string, arguments []usecaseifs.IOSCMessageArgument) *Message {
	return &Message{address: address, arguments: arguments}
}

func (m *Message) GetAddress() string {
	return m.address
}

func (m *Message) GetArguments() []usecaseifs.IOSCMessageArgument {
	return m.arguments
}

func (m *Message) Equal(msg usecaseifs.IOSCMessage) bool {
	if m.GetAddress() != msg.GetAddress() {
		return false
	}

	if len(m.arguments) != len(msg.GetArguments()) {
		return false
	}

	for i, a := range m.arguments {
		if a.GetType() != msg.GetArguments()[i].GetType() {
			return false
		}
		if a.GetValue() != msg.GetArguments()[i].GetValue() {
			return false
		}
	}

	return true
}

func (m *Message) String() string {
	argStrings := []string{}
	for _, arg := range m.arguments {
		argStrings = append(argStrings, arg.String())
	}
	return fmt.Sprintf("Message(address: %s, arguments: [%s])", m.address, strings.Join(argStrings, ", "))
}
