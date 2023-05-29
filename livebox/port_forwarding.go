package livebox

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// PortForwarding describes the configuration of a port forwarding rule.
type PortForwarding struct {
	Name         string
	Protocol     Protocol
	ExternalPort int
	InternalPort int
	PortRange    int
	Destination  string
	Enabled      bool
}

type getPortForwardingResp struct {
	ID                    string `json:"Id"`
	Origin                string `json:"Origin"`
	Description           string `json:"Description"`
	Status                string `json:"Status"`
	SourceInterface       string `json:"SourceInterface"`
	Protocol              string `json:"Protocol"`
	ExternalPort          string `json:"ExternalPort"`
	InternalPort          string `json:"InternalPort"`
	SourcePrefix          string `json:"SourcePrefix"`
	DestinationIPAddress  string `json:"DestinationIPAddress"`
	DestinationMACAddress string `json:"DestinationMACAddress"`
	LeaseDuration         int    `json:"LeaseDuration"`
	HairpinNAT            bool   `json:"HairpinNAT"`
	SymmetricSNAT         bool   `json:"SymmetricSNAT"`
	UPnPV1Compat          bool   `json:"UPnPV1Compat"`
	Enable                bool   `json:"Enable"`
}

// ListPortForwardings returns all the port forwarding rules currently configured.
func (c *Client) ListPortForwardings() ([]PortForwarding, error) {
	payload := &apiRequest{
		Service: "Firewall",
		Method:  "getPortForwarding",
		Parameters: map[string]any{
			"origin": "webui",
		},
	}

	data, err := c.doReq(payload)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	var pfr map[string]getPortForwardingResp
	if err = json.Unmarshal(data, &pfr); err != nil {
		return nil, fmt.Errorf("unmarshal data: %w", err)
	}

	out := make([]PortForwarding, 0, len(pfr))
	for id, raw := range pfr {
		externalPort, portRange, err := parsePortRange(raw.ExternalPort)
		if err != nil {
			return nil, fmt.Errorf("parse port range: %w", err)
		}

		internalPort, err := strconv.Atoi(raw.InternalPort)
		if err != nil {
			return nil, fmt.Errorf("parse internal port: %w", err)
		}

		pf := PortForwarding{
			Name:         strings.TrimPrefix(id, "webui_"),
			Protocol:     parseProtocol(raw.Protocol),
			ExternalPort: externalPort,
			InternalPort: internalPort,
			PortRange:    portRange,
			Destination:  raw.DestinationIPAddress,
			Enabled:      raw.Enable,
		}
		out = append(out, pf)
	}

	return out, nil
}

// GetPortForwarding returns the port forwarding rule configured matching the given name, if found.
// This method is just implemented for convenience, as the Livebox API does not seem to expose an endpoint
// to retrieve a single port forwarding rule, so it lists them all and filters the result.
func (c *Client) GetPortForwarding(name string) (*PortForwarding, error) {
	pfs, err := c.ListPortForwardings()
	if err != nil {
		return nil, fmt.Errorf("list port forwardings: %w", err)
	}

	for _, pf := range pfs {
		if pf.Name == name {
			return &pf, nil
		}
	}

	return nil, errors.New("port forward not found")
}

// PortForwardingConfig configures a port forwarding rule.
type PortForwardingConfig struct {
	Name         string
	ExternalPort int
	InternalPort int
	PortRange    int
	Protocol     Protocol
	Destination  string
	Enabled      bool
}

// validate performs some basic validation on a port forward rule configuration.
func (c PortForwardingConfig) validate() error {
	if c.Name == "" {
		return errors.New("empty name")
	}

	if c.ExternalPort < 1 || c.ExternalPort > 65535 {
		return errors.New("invalid external port; must be between 1 and 65535")
	}

	if c.InternalPort < 1 || c.InternalPort > 65535 {
		return errors.New("invalid external port; must be between 1 and 65535")
	}

	if c.PortRange < 0 || c.PortRange > 65535 {
		return errors.New("invalid port range; must be between 1 and 65535")
	}

	if c.Protocol != ProtocolTCP && c.Protocol != ProtocolUDP && c.Protocol != ProtocolTCPUDP {
		return fmt.Errorf("invalid protocol; must be one of: %q, %q or %q", "tcp", "udp", "tcp/udp")
	}

	if ip := net.ParseIP(c.Destination); ip == nil {
		return errors.New("invalid destination; must be a valid IP address")
	}

	return nil
}

// UpsertPortForwarding upserts the given port forwarding rule.
func (c *Client) UpsertPortForwarding(cfg PortForwardingConfig) error {
	if err := cfg.validate(); err != nil {
		return fmt.Errorf("validate configuration: %w", err)
	}

	payload := &apiRequest{
		Service: "Firewall",
		Method:  "setPortForwarding",
		Parameters: map[string]any{
			"id":                   "webui_" + cfg.Name,
			"description":          cfg.Name,
			"protocol":             cfg.Protocol.intString(),
			"internalPort":         cfg.InternalPort,
			"externalPort":         formatPortRange(cfg.ExternalPort, cfg.PortRange),
			"destinationIPAddress": cfg.Destination,
			"sourcePrefix":         "",
			"persistent":           true,
			"enable":               cfg.Enabled,
			"sourceInterface":      "data",
			"origin":               "webui",
		},
	}

	if _, err := c.doReq(payload); err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	return nil
}

// DeletePortForwarding deletes the port forwarding rule matching the given name.
func (c *Client) DeletePortForwarding(name string) error {
	payload := &apiRequest{
		Service: "Firewall",
		Method:  "deletePortForwarding",
		Parameters: map[string]any{
			"id":     "webui_" + name,
			"origin": "webui",
		},
	}

	if _, err := c.doReq(payload); err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	return nil
}

// parsePortRange parses a port expressed as a string, with an optional range written
// using the following syntax: "10000-10005".
func parsePortRange(port string) (externalPort int, portRange int, err error) {
	parts := strings.SplitN(port, "-", 2)
	externalPortStart, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse external port: %w", err)
	}

	// We only have one part, there is no range specified.
	if len(parts) == 1 {
		return externalPortStart, 0, nil
	}

	externalPortEnd, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse external port range: %w", err)
	}

	return externalPortStart, externalPortEnd - externalPortStart, nil
}

// formatPortRange does the opposite of parsePortRange. It returns a port and an optional range as a string
// using the following syntax: "10000-10005". If the range is 0 or 1, then the port is simply converted to a string.
func formatPortRange(port, portRange int) string {
	if portRange == 0 || portRange == 1 {
		return strconv.Itoa(port)
	}

	return fmt.Sprintf("%s-%s", strconv.Itoa(port), strconv.Itoa(port+portRange))
}
