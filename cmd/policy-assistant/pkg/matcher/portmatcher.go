package matcher

import (
	"encoding/json"
	"fmt"
	"sort"

	collectionsjson "github.com/mattfenwick/collections/pkg/json"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type PortMatcher interface {
	Matches(portInt int, portName string, protocol v1.Protocol) bool
	GetPrimaryKey() string
}

type AllPortMatcher struct{}

func (ap *AllPortMatcher) Matches(portInt int, portName string, protocol v1.Protocol) bool {
	return true
}

func (ap *AllPortMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all ports",
	})
}

func (ap *AllPortMatcher) GetPrimaryKey() string {
	return "all ports"
}

// PortProtocolMatcher models a matcher based on:
// 1. Protocol
// 2. Either a) port number or b) port name.
// This matcher is modeled after the confusing behavior of v1 NetPol's Port field.
// We map NetworkPolicy v2 to this matcher despite v2's better approach to distinguishing named ports from a port number.
// This matcher requires that you specify the Protocol for a NamedPort.
// In NetPol v1, there is the pathological case where the user can define a NetworkPolicy rule allowing a NamedPort on the wrong Protocol.
type PortProtocolMatcher struct {
	// Port is either a port number or a port name.
	// If nil, all ports are matched.
	Port *intstr.IntOrString
	// Protocol must be set, even for a named port
	Protocol v1.Protocol
}

// MatchesPortProtocol does not implement the PortMatcher interface, purposely!
func (p *PortProtocolMatcher) MatchesPortProtocol(portInt int, portName string, protocol v1.Protocol) bool {
	if p.Port != nil {
		return isPortMatch(*p.Port, portInt, portName) && p.Protocol == protocol
	}
	return p.Protocol == protocol
}

func (p *PortProtocolMatcher) Equals(other *PortProtocolMatcher) bool {
	if p.Protocol != other.Protocol {
		return false
	}
	if p.Port == nil && other.Port == nil {
		return true
	}
	if (p.Port == nil && other.Port != nil) || (p.Port != nil && other.Port == nil) {
		return false
	}
	return isIntStringEqual(*p.Port, *other.Port)
}

func (p *PortProtocolMatcher) GetPrimaryKey() string {
	return fmt.Sprintf("Type: %s, Port: %s, Protocol: %s", "Port Protocol", p.Port.String(), p.Protocol)
}

// PortRangeMatcher works with endports to specify a range of matched numeric ports.
type PortRangeMatcher struct {
	From     int
	To       int
	Protocol v1.Protocol
}

func (prm *PortRangeMatcher) MatchesPortProtocol(portInt int, protocol v1.Protocol) bool {
	return prm.From <= portInt && portInt <= prm.To && prm.Protocol == protocol
}

func (prm *PortRangeMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":     "port range",
		"From":     prm.From,
		"To":       prm.To,
		"Protocol": prm.Protocol,
	})
}

func (prm *PortRangeMatcher) GetPrimaryKey() string {
	return fmt.Sprintf("Type: %s, From: %d, To: %d, Protocol: %s", "port range", prm.From, prm.To, prm.Protocol)
}

// SpecificPortMatcher models the case where traffic must match a named or numbered port
type SpecificPortMatcher struct {
	Ports      []*PortProtocolMatcher
	PortRanges []*PortRangeMatcher
}

func (s *SpecificPortMatcher) Matches(portInt int, portName string, protocol v1.Protocol) bool {
	for _, matcher := range s.Ports {
		if matcher.MatchesPortProtocol(portInt, portName, protocol) {
			return true
		}
	}
	for _, matcher := range s.PortRanges {
		if matcher.MatchesPortProtocol(portInt, protocol) {
			return true
		}
	}
	return false
}

func (s *SpecificPortMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":       "specific ports",
		"Ports":      s.Ports,
		"PortRanges": s.PortRanges,
	})
}

func (s *SpecificPortMatcher) GetPrimaryKey() string {
	var p string
	for _, v := range s.Ports {
		p += v.GetPrimaryKey()
	}

	var pr string
	for _, v := range s.PortRanges {
		pr += v.GetPrimaryKey()
	}

	return fmt.Sprintf("Type: %s, Ports: %s, PortRanges: %s", "specific port", p, pr)
}

func (s *SpecificPortMatcher) Combine(other *SpecificPortMatcher) *SpecificPortMatcher {
	logrus.Debugf("SpecificPortMatcher Combined:\n%s\n", collectionsjson.MustMarshalToString([]interface{}{s, other}))

	pps := append([]*PortProtocolMatcher{}, s.Ports...)
	for _, otherPP := range other.Ports {
		foundDuplicate := false
		for _, pp := range pps {
			if pp.Equals(otherPP) {
				foundDuplicate = true
				break
			}
		}
		if !foundDuplicate {
			pps = append(pps, otherPP)
		}
	}
	sort.Slice(pps, func(i, j int) bool {
		// first, run it forward
		if isPortLessThan(pps[i].Port, pps[j].Port) {
			return true
		}
		// flip it around, run it the other way
		if isPortLessThan(pps[j].Port, pps[i].Port) {
			return false
		}
		// neither is less than the other?  fall back to protocol
		return pps[i].Protocol < pps[j].Protocol
	})

	// TODO compact port ranges
	ranges := append(s.PortRanges, other.PortRanges...)
	// TODO sort port ranges

	logrus.Debugf("ports:\n%s\n", collectionsjson.MustMarshalToString(pps))

	return &SpecificPortMatcher{Ports: pps, PortRanges: ranges}
}

func (s *SpecificPortMatcher) Subtract(other *SpecificPortMatcher) (bool, *SpecificPortMatcher) {
	// TODO actually subtract ranges
	remainingRanges := s.PortRanges

	var remaining []*PortProtocolMatcher
	for _, thisPort := range s.Ports {
		found := false
		for _, otherPort := range other.Ports {
			if thisPort.Equals(otherPort) {
				found = true
				break
			}
		}
		if !found {
			remaining = append(remaining, thisPort)
		}
	}
	if len(remainingRanges) == 0 && len(remaining) == 0 {
		return true, nil
	}
	return false, &SpecificPortMatcher{Ports: remaining, PortRanges: remainingRanges}
}

// isPortLessThan orders from low to high:
// - nil
// - string
// - int
func isPortLessThan(a *intstr.IntOrString, b *intstr.IntOrString) bool {
	if a == nil {
		return b != nil
	}
	if b == nil {
		return false
	}
	switch a.Type {
	case intstr.Int:
		switch b.Type {
		case intstr.Int:
			return a.IntVal < b.IntVal
		case intstr.String:
			return false
		default:
			panic("invalid type")
		}
	case intstr.String:
		switch b.Type {
		case intstr.Int:
			return true
		case intstr.String:
			return a.StrVal < b.StrVal
		default:
			panic("invalid type")
		}
	default:
		panic("invalid type")
	}
}

func isPortMatch(a intstr.IntOrString, portInt int, portName string) bool {
	switch a.Type {
	case intstr.Int:
		return int(a.IntVal) == portInt
	case intstr.String:
		return a.StrVal == portName
	default:
		panic("invalid type")
	}
}

func isIntStringEqual(a intstr.IntOrString, b intstr.IntOrString) bool {
	switch a.Type {
	case intstr.Int:
		switch b.Type {
		case intstr.Int:
			return a.IntVal == b.IntVal
		case intstr.String:
			return false
		default:
			panic("invalid type")
		}
	case intstr.String:
		switch b.Type {
		case intstr.Int:
			return false
		case intstr.String:
			return a.StrVal == b.StrVal
		default:
			panic("invalid type")
		}
	default:
		panic("invalid type")
	}
}
