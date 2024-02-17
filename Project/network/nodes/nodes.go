package nodes

import (
	"fmt"
	"net"
	"network/conn"
	"time"
)

type NetworkNodeRegistry struct {
	Nodes []string
	New   string
	Lost  []string
}

const interval = 15 * time.Millisecond
const timeout = 500 * time.Millisecond

func Client(port int, id string, enable <-chan bool) {
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255%d", port))

	for {
		select {
		case transmitData := <-enable:
			if transmitData {
				conn.WriteTo([]byte(id), addr)
			}
		case <-time.After(interval):
		}
	}
}

func Server(port int, updateChannel <-chan NetworkNodeRegistry) {
	//
}
