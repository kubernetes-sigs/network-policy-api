package probe

import "github.com/pkg/errors"

type Connectivity string

const (
	ConnectivityUnknown             Connectivity = "unknown"
	ConnectivityCheckFailed         Connectivity = "checkfailed"
	ConnectivityInvalidNamedPort    Connectivity = "invalidnamedport"
	ConnectivityInvalidPortProtocol Connectivity = "invalidportprotocol"
	ConnectivityBlocked             Connectivity = "blocked"
	ConnectivityAllowed             Connectivity = "allowed"
	// ConnectivityUndefined e.g. for loopback traffic
	ConnectivityUndefined Connectivity = "undefined"
)

var AllConnectivity = []Connectivity{
	ConnectivityUnknown,
	ConnectivityCheckFailed,
	ConnectivityInvalidNamedPort,
	ConnectivityInvalidPortProtocol,
	ConnectivityBlocked,
	ConnectivityAllowed,
}

func (p Connectivity) ShortString() string {
	switch p {
	case ConnectivityUnknown:
		return "?"
	case ConnectivityCheckFailed:
		return "!"
	case ConnectivityBlocked:
		return "X"
	case ConnectivityAllowed:
		return "."
	case ConnectivityInvalidNamedPort:
		return "P"
	case ConnectivityInvalidPortProtocol:
		return "N"
	case ConnectivityUndefined:
		return "#"
	default:
		panic(errors.Errorf("invalid Connectivity value: %+v", p))
	}
}
