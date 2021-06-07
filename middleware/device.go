package middleware

import "net"

type DeviceMode int

const (
	SERVER DeviceMode = iota
	CLIENT
)

type Device interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Mode() DeviceMode
	Name() string
}
