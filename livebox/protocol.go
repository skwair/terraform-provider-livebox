package livebox

// Protocol used for port forwarding.
type Protocol string

// List of supported protocols:
const (
	ProtocolUnknown Protocol = "unknown"
	ProtocolTCP     Protocol = "tcp"
	ProtocolUDP     Protocol = "udp"
	ProtocolTCPUDP  Protocol = "tcp/udp"
)

func (p Protocol) intString() string {
	switch p {
	case ProtocolTCP:
		return "6"
	case ProtocolUDP:
		return "17"
	case ProtocolTCPUDP:
		return "6,17"
	default:
		return ""
	}
}

func parseProtocol(p string) Protocol {
	switch p {
	case "6":
		return ProtocolTCP
	case "17":
		return ProtocolUDP
	case "6,17":
		return ProtocolTCPUDP
	default:
		return ProtocolUnknown
	}
}
