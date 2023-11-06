package entities

import (
	"fmt"
	"strings"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IOSCMessage = &PrefixedOSCMessage{}

// PrefixedOSCMessage is a wrapper to add a prefix for the address of a message.
type PrefixedOSCMessage struct {
	Prefix  string
	Message usecaseifs.IOSCMessage
}

func NewPrefixedOSCMessage(prefix string, message usecaseifs.IOSCMessage) *PrefixedOSCMessage {
	return &PrefixedOSCMessage{Prefix: prefix, Message: message}
}

func (p PrefixedOSCMessage) Equal(msg usecaseifs.IOSCMessage) bool {
	return p.Message.Equal(msg)
}

func (p PrefixedOSCMessage) GetAddress() string {
	if p.Prefix == "" {
		return p.Message.GetAddress()
	}
	return fmt.Sprintf("%s/%s", p.Prefix, p.Message.GetAddress())
}

func (p PrefixedOSCMessage) GetArguments() []usecaseifs.IOSCMessageArgument {
	return p.Message.GetArguments()
}

func (p PrefixedOSCMessage) String() string {
	argStrings := []string{}
	for _, arg := range p.GetArguments() {
		argStrings = append(argStrings, arg.String())
	}
	return fmt.Sprintf("Message(address: %s, arguments: [%s])", p.GetAddress(), strings.Join(argStrings, ", "))
}
