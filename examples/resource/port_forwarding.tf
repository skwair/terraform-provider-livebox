resource "livebox_port_forwarding" "wireguard" {
  name = "wireguard"
  protocol = "udp"
  external_port = 51820
  internal_port = 51820
  destination = "192.168.10.200"
  enabled = true
}
