package entities

import "net.kopias.oscbridge/app/usecase/usecaseifs"

// OscConnectionDetails wraps a connection and adds metadata.
type OscConnectionDetails struct {
	// Name is the name of this connection.
	Name string

	// Prefix determines the address prefix for this connection.
	Prefix string

	Connection usecaseifs.IOSCConnection
}

func NewOscConnectionDetails(name string, prefix string, connection usecaseifs.IOSCConnection) *OscConnectionDetails {
	return &OscConnectionDetails{Name: name, Prefix: prefix, Connection: connection}
}
