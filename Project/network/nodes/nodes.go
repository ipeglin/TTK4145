package nodes

import (
	"fmt"
	"net"
	"network/conn"
	"sort"
	"time"
)

type NetworkNodeRegistry struct {
	Nodes []string
	New   string
	Lost  []string
}

const interval = 150 * time.Millisecond
const timeout = 3000 * time.Millisecond

func Sender(port int, id string, enableTransmit <-chan bool) {
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", port))
	enable := true

	for {
		select {
		case enable = <-enableTransmit:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Receiver(port int, updateChannel chan<- NetworkNodeRegistry) {
	var buffer [1024]byte
	var reg NetworkNodeRegistry
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buffer[0:])

		id := string(buffer[:n])

		reg.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				reg.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		reg.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				reg.Lost = append(reg.Lost, k)
				delete(lastSeen, k)
			}
		}

		if updated {
			reg.Nodes = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				reg.Nodes = append(reg.Nodes, k)
			}

			sort.Strings(reg.Nodes)
			updateChannel <- reg
		}
	}
}
