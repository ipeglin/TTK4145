package local

import (
	"net"
	"strings"
)

var hostIP string

func GetIP() (string, error) {
	if hostIP == "" {
		conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()

		hostIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return hostIP, nil
}
