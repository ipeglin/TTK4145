//go:build linux
// +build linux

package conn

import (
	"net"
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
)

func DialBroadcastUDP(port int) net.PacketConn {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		logrus.Error("Error: Socket:", err)
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		logrus.Error("Error: SetSockOpt REUSEADDR:", err)
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil {
		logrus.Error("Error: SetSockOpt BROADCAST:", err)
	}
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})
	if err != nil {
		logrus.Error("Error: Bind:", err)
	}

	f := os.NewFile(uintptr(s), "")
	conn, err := net.FilePacketConn(f)
	if err != nil {
		logrus.Error("Error: FilePacketConn:", err)
	}
	f.Close()

	return conn
}
