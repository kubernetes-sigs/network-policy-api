package matcher

import (
	"encoding/json"

	v1 "k8s.io/api/core/v1"
)

var (
	AllPeersPorts = &AllPeersMatcher{}
)

/*
PeerMatcher matches a peer against an ANP, BANP, or v1 NetPol rule.

These are the original PeerMatcher implementations made for v1 NetPol:
- AllPeersMatcher
- PortsForAllPeersMatcher
- IPPeerMatcher
- PodPeerMatcher

Now we also have PeerMatcherV2, a wrapper for the above to model ANP and BANP,
as well as NamespaceMatcher objects for SameLabels and NotSameLabels.

All of these (except AllPeersMatcher) use a PortMatcher.
If the traffic doesn't match the port matcher, then Matches() will be false.
*/
type PeerMatcher interface {
	Matches(subject, peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool
}

// AllPeerMatcher matches all pod to pod traffic.
type AllPeersMatcher struct{}

func (a *AllPeersMatcher) Matches(_, peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	return true
}

func (a *AllPeersMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all peers",
	})
}

// AllPeerMatcher matches all pod to pod traffic that satisfies its port matcher
type PortsForAllPeersMatcher struct {
	Port PortMatcher
}

func (p *PortsForAllPeersMatcher) Matches(_, peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	return p.Port.Matches(portInt, portName, protocol)
}

func (p *PortsForAllPeersMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all peers for port",
		"Port": p.Port,
	})
}
